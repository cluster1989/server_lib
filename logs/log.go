// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package logs provide a general log interface
// Usage:
//
// import "github.com/astaxie/beego/logs"
//
//	log := NewLogger(10000)
//	log.SetLogger("console", "")
//
//	> the first params stand for how many channel
//
// Use it like this:
//
//	log.Trace("trace")
//	log.Info("info")
//	log.Warn("warning")
//	log.Debug("debug")
//	log.Critical("critical")
//
//  more docs http://beego.me/docs/module/logs.md
package logs

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RFC5424 log message levels.
const (
	LevelEmergency = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

// levelLogLogger is defined to implement log.Logger
// the real log level will be LevelEmergency
const levelLoggerImpl = -1

// Name for adapter with beego official support
const (
	AdapterConsole   = "console"
	AdapterFile      = "file"
	AdapterMultiFile = "multifile"
	AdapterMail      = "smtp"
	AdapterConn      = "conn"
)

// Legacy log level constants to ensure backwards compatibility.
const (
	LevelInfo  = LevelInformational
	LevelTrace = LevelDebug
	LevelWarn  = LevelWarning
)

type newLoggerFunc func() Logger

// Logger defines the behavior of a log provider.
type Logger interface {
	Init(config string) error
	WriteMsg(when time.Time, msg string, level int) error
	Destroy()
	Flush()
}

var adapters = make(map[string]newLoggerFunc)
var levelPrefix = [LevelDebug + 1]string{"[M] ", "[A] ", "[C] ", "[E] ", "[W] ", "[N] ", "[I] ", "[D] "}

// Register makes a log provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, log newLoggerFunc) {
	if log == nil {
		panic("logs: Register provide is nil")
	}
	if _, dup := adapters[name]; dup {
		panic("logs: Register called twice for provider " + name)
	}
	adapters[name] = log
}

// LibLogger is default logger in beego application.
// it can contain several providers and log message into all providers.
type LibLogger struct {
	lock                sync.Mutex
	level               int
	init                bool
	enableFuncCallDepth bool
	loggerFuncCallDepth int
	asynchronous        bool
	msgChanLen          int64
	msgChan             chan *logMsg
	signalChan          chan string
	wg                  sync.WaitGroup
	outputs             []*nameLogger
}

const defaultAsyncMsgLen = 1e3

type nameLogger struct {
	Logger
	name string
}

type logMsg struct {
	level int
	msg   string
	when  time.Time
}

var logMsgPool *sync.Pool

// NewLogger returns a new LibLogger.
// channelLen means the number of messages in chan(used where asynchronous is true).
// if the buffering chan is full, logger adapters write to file or other way.
func NewLogger(channelLens ...int64) *LibLogger {
	libLog := new(LibLogger)
	libLog.level = LevelDebug
	libLog.loggerFuncCallDepth = 2
	libLog.msgChanLen = append(channelLens, 0)[0]
	if libLog.msgChanLen <= 0 {
		libLog.msgChanLen = defaultAsyncMsgLen
	}
	libLog.signalChan = make(chan string, 1)
	libLog.setLogger(AdapterConsole)
	return libLog
}

// Async set the log to asynchronous and start the goroutine
func (libLog *LibLogger) Async(msgLen ...int64) *LibLogger {
	libLog.lock.Lock()
	defer libLog.lock.Unlock()
	if libLog.asynchronous {
		return libLog
	}
	libLog.asynchronous = true
	if len(msgLen) > 0 && msgLen[0] > 0 {
		libLog.msgChanLen = msgLen[0]
	}
	libLog.msgChan = make(chan *logMsg, libLog.msgChanLen)
	logMsgPool = &sync.Pool{
		New: func() interface{} {
			return &logMsg{}
		},
	}
	libLog.wg.Add(1)
	go libLog.startLogger()
	return libLog
}

// SetLogger provides a given logger adapter into LibLogger with config string.
// config need to be correct JSON as string: {"interval":360}.
func (libLog *LibLogger) setLogger(adapterName string, configs ...string) error {
	config := append(configs, "{}")[0]
	for _, l := range libLog.outputs {
		if l.name == adapterName {
			return fmt.Errorf("logs: duplicate adaptername %q (you have set this logger before)", adapterName)
		}
	}

	log, ok := adapters[adapterName]
	if !ok {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adapterName)
	}

	lg := log()
	err := lg.Init(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "logs.LibLogger.SetLogger: "+err.Error())
		return err
	}
	libLog.outputs = append(libLog.outputs, &nameLogger{name: adapterName, Logger: lg})
	return nil
}

// SetLogger provides a given logger adapter into LibLogger with config string.
// config need to be correct JSON as string: {"interval":360}.
func (libLog *LibLogger) SetLogger(adapterName string, configs ...string) error {
	libLog.lock.Lock()
	defer libLog.lock.Unlock()
	if !libLog.init {
		libLog.outputs = []*nameLogger{}
		libLog.init = true
	}
	return libLog.setLogger(adapterName, configs...)
}

// DelLogger remove a logger adapter in LibLogger.
func (libLog *LibLogger) DelLogger(adapterName string) error {
	libLog.lock.Lock()
	defer libLog.lock.Unlock()
	outputs := []*nameLogger{}
	for _, lg := range libLog.outputs {
		if lg.name == adapterName {
			lg.Destroy()
		} else {
			outputs = append(outputs, lg)
		}
	}
	if len(outputs) == len(libLog.outputs) {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adapterName)
	}
	libLog.outputs = outputs
	return nil
}

func (libLog *LibLogger) writeToLoggers(when time.Time, msg string, level int) {
	for _, l := range libLog.outputs {
		err := l.WriteMsg(when, msg, level)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to WriteMsg to adapter:%v,error:%v\n", l.name, err)
		}
	}
}

func (libLog *LibLogger) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	// writeMsg will always add a '\n' character
	if p[len(p)-1] == '\n' {
		p = p[0 : len(p)-1]
	}
	// set levelLoggerImpl to ensure all log message will be write out
	err = libLog.writeMsg(levelLoggerImpl, string(p))
	if err == nil {
		return len(p), err
	}
	return 0, err
}

func (libLog *LibLogger) writeMsg(logLevel int, msg string, v ...interface{}) error {
	if !libLog.init {
		libLog.lock.Lock()
		libLog.setLogger(AdapterConsole)
		libLog.lock.Unlock()
	}
	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}
	when := time.Now()
	if libLog.enableFuncCallDepth {
		_, file, line, ok := runtime.Caller(libLog.loggerFuncCallDepth)
		if !ok {
			file = "???"
			line = 0
		}
		_, filename := path.Split(file)

		msg = "[" + filename + ":" + strconv.Itoa(line) + "] " + msg
	}

	//set level info in front of filename info
	if logLevel == levelLoggerImpl {
		// set to emergency to ensure all log will be print out correctly
		logLevel = LevelEmergency
	} else {
		msg = levelPrefix[logLevel] + msg
	}

	if libLog.asynchronous {
		lm := logMsgPool.Get().(*logMsg)
		lm.level = logLevel
		lm.msg = msg
		lm.when = when
		libLog.msgChan <- lm
	} else {
		libLog.writeToLoggers(when, msg, logLevel)
	}
	return nil
}

// SetLevel Set log message level.
// If message level (such as LevelDebug) is higher than logger level (such as LevelWarning),
// log providers will not even be sent the message.
func (libLog *LibLogger) SetLevel(l int) {
	libLog.level = l
}

// SetLogFuncCallDepth set log funcCallDepth
func (libLog *LibLogger) SetLogFuncCallDepth(d int) {
	libLog.loggerFuncCallDepth = d
}

// GetLogFuncCallDepth return log funcCallDepth for wrapper
func (libLog *LibLogger) GetLogFuncCallDepth() int {
	return libLog.loggerFuncCallDepth
}

// EnableFuncCallDepth enable log funcCallDepth
func (libLog *LibLogger) EnableFuncCallDepth(b bool) {
	libLog.enableFuncCallDepth = b
}

// start logger chan reading.
// when chan is not empty, write logs.
func (libLog *LibLogger) startLogger() {
	gameOver := false
	for {
		select {
		case bm := <-libLog.msgChan:
			libLog.writeToLoggers(bm.when, bm.msg, bm.level)
			logMsgPool.Put(bm)
		case sg := <-libLog.signalChan:
			// Now should only send "flush" or "close" to libLog.signalChan
			libLog.flush()
			if sg == "close" {
				for _, l := range libLog.outputs {
					l.Destroy()
				}
				libLog.outputs = nil
				gameOver = true
			}
			libLog.wg.Done()
		}
		if gameOver {
			break
		}
	}
}

// Emergency Log EMERGENCY level message.
func (libLog *LibLogger) Emergency(format string, v ...interface{}) {
	if LevelEmergency > libLog.level {
		return
	}
	libLog.writeMsg(LevelEmergency, format, v...)
}

// Alert Log ALERT level message.
func (libLog *LibLogger) Alert(format string, v ...interface{}) {
	if LevelAlert > libLog.level {
		return
	}
	libLog.writeMsg(LevelAlert, format, v...)
}

// Critical Log CRITICAL level message.
func (libLog *LibLogger) Critical(format string, v ...interface{}) {
	if LevelCritical > libLog.level {
		return
	}
	libLog.writeMsg(LevelCritical, format, v...)
}

// Error Log ERROR level message.
func (libLog *LibLogger) Error(format string, v ...interface{}) {
	if LevelError > libLog.level {
		return
	}
	libLog.writeMsg(LevelError, format, v...)
}

// Warning Log WARNING level message.
func (libLog *LibLogger) Warning(format string, v ...interface{}) {
	if LevelWarn > libLog.level {
		return
	}
	libLog.writeMsg(LevelWarn, format, v...)
}

// Notice Log NOTICE level message.
func (libLog *LibLogger) Notice(format string, v ...interface{}) {
	if LevelNotice > libLog.level {
		return
	}
	libLog.writeMsg(LevelNotice, format, v...)
}

// Informational Log INFORMATIONAL level message.
func (libLog *LibLogger) Informational(format string, v ...interface{}) {
	if LevelInfo > libLog.level {
		return
	}
	libLog.writeMsg(LevelInfo, format, v...)
}

// Debug Log DEBUG level message.
func (libLog *LibLogger) Debug(format string, v ...interface{}) {
	if LevelDebug > libLog.level {
		return
	}
	libLog.writeMsg(LevelDebug, format, v...)
}

// Warn Log WARN level message.
// compatibility alias for Warning()
func (libLog *LibLogger) Warn(format string, v ...interface{}) {
	if LevelWarn > libLog.level {
		return
	}
	libLog.writeMsg(LevelWarn, format, v...)
}

// Info Log INFO level message.
// compatibility alias for Informational()
func (libLog *LibLogger) Info(format string, v ...interface{}) {
	if LevelInfo > libLog.level {
		return
	}
	libLog.writeMsg(LevelInfo, format, v...)
}

// Trace Log TRACE level message.
// compatibility alias for Debug()
func (libLog *LibLogger) Trace(format string, v ...interface{}) {
	if LevelDebug > libLog.level {
		return
	}
	libLog.writeMsg(LevelDebug, format, v...)
}

// Flush flush all chan data.
func (libLog *LibLogger) Flush() {
	if libLog.asynchronous {
		libLog.signalChan <- "flush"
		libLog.wg.Wait()
		libLog.wg.Add(1)
		return
	}
	libLog.flush()
}

// Close close logger, flush all chan data and destroy all adapters in LibLogger.
func (libLog *LibLogger) Close() {
	if libLog.asynchronous {
		libLog.signalChan <- "close"
		libLog.wg.Wait()
		close(libLog.msgChan)
	} else {
		libLog.flush()
		for _, l := range libLog.outputs {
			l.Destroy()
		}
		libLog.outputs = nil
	}
	close(libLog.signalChan)
}

// Reset close all outputs, and set libLog.outputs to nil
func (libLog *LibLogger) Reset() {
	libLog.Flush()
	for _, l := range libLog.outputs {
		l.Destroy()
	}
	libLog.outputs = nil
}

func (libLog *LibLogger) flush() {
	if libLog.asynchronous {
		for {
			if len(libLog.msgChan) > 0 {
				bm := <-libLog.msgChan
				libLog.writeToLoggers(bm.when, bm.msg, bm.level)
				logMsgPool.Put(bm)
				continue
			}
			break
		}
	}
	for _, l := range libLog.outputs {
		l.Flush()
	}
}

// libLogger references the used application logger.
var libLogger = NewLogger()

// GetLibLogger returns the default LibLogger
func GetLibLogger() *LibLogger {
	return libLogger
}

var libLoggerMap = struct {
	sync.RWMutex
	logs map[string]*log.Logger
}{
	logs: map[string]*log.Logger{},
}

// GetLogger returns the default LibLogger
func GetLogger(prefixes ...string) *log.Logger {
	prefix := append(prefixes, "")[0]
	if prefix != "" {
		prefix = fmt.Sprintf(`[%s] `, strings.ToUpper(prefix))
	}
	libLoggerMap.RLock()
	l, ok := libLoggerMap.logs[prefix]
	if ok {
		libLoggerMap.RUnlock()
		return l
	}
	libLoggerMap.RUnlock()
	libLoggerMap.Lock()
	defer libLoggerMap.Unlock()
	l, ok = libLoggerMap.logs[prefix]
	if !ok {
		l = log.New(libLogger, prefix, 0)
		libLoggerMap.logs[prefix] = l
	}
	return l
}

// Reset will remove all the adapter
func Reset() {
	libLogger.Reset()
}

// Async set the beelogger with Async mode and hold msglen messages
func Async(msgLen ...int64) *LibLogger {
	return libLogger.Async(msgLen...)
}

// SetLevel sets the global log level used by the simple logger.
func SetLevel(l int) {
	libLogger.SetLevel(l)
}

// EnableFuncCallDepth enable log funcCallDepth
func EnableFuncCallDepth(b bool) {
	libLogger.enableFuncCallDepth = b
}

// SetLogFuncCall set the CallDepth, default is 4
func SetLogFuncCall(b bool) {
	libLogger.EnableFuncCallDepth(b)
	libLogger.SetLogFuncCallDepth(4)
}

// SetLogFuncCallDepth set log funcCallDepth
func SetLogFuncCallDepth(d int) {
	libLogger.loggerFuncCallDepth = d
}

// SetLogger sets a new logger.
func SetLogger(adapter string, config ...string) error {
	return libLogger.SetLogger(adapter, config...)
}

// Emergency logs a message at emergency level.
func Emergency(f interface{}, v ...interface{}) {
	libLogger.Emergency(formatLog(f, v...))
}

// Alert logs a message at alert level.
func Alert(f interface{}, v ...interface{}) {
	libLogger.Alert(formatLog(f, v...))
}

// Critical logs a message at critical level.
func Critical(f interface{}, v ...interface{}) {
	libLogger.Critical(formatLog(f, v...))
}

// Error logs a message at error level.
func Error(f interface{}, v ...interface{}) {
	libLogger.Error(formatLog(f, v...))
}

// Warning logs a message at warning level.
func Warning(f interface{}, v ...interface{}) {
	libLogger.Warn(formatLog(f, v...))
}

// Warn compatibility alias for Warning()
func Warn(f interface{}, v ...interface{}) {
	libLogger.Warn(formatLog(f, v...))
}

// Notice logs a message at notice level.
func Notice(f interface{}, v ...interface{}) {
	libLogger.Notice(formatLog(f, v...))
}

// Informational logs a message at info level.
func Informational(f interface{}, v ...interface{}) {
	libLogger.Info(formatLog(f, v...))
}

// Info compatibility alias for Warning()
func Info(f interface{}, v ...interface{}) {
	libLogger.Info(formatLog(f, v...))
}

// Debug logs a message at debug level.
func Debug(f interface{}, v ...interface{}) {
	libLogger.Debug(formatLog(f, v...))
}

// Trace logs a message at trace level.
// compatibility alias for Warning()
func Trace(f interface{}, v ...interface{}) {
	libLogger.Trace(formatLog(f, v...))
}

func formatLog(f interface{}, v ...interface{}) string {
	var msg string
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if strings.Contains(msg, "%") && !strings.Contains(msg, "%%") {
			//format string
		} else {
			//do not contain format char
			msg += strings.Repeat(" %v", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}
