package logs2

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"
	"time"
)

// 日志的接口
type Logger2 interface {
	// 初始化日志服务
	Init(config interface{}) error
	// 写日志
	WriteMsg(when time.Time, msg string, level int) error
	// 摧毁
	Destroy()
	// 一次写完
	Flush()
}

var defaultLogMsgPool *sync.Pool

var levelPrefix = [LogLevelDebug + 1]string{"[M] ", "[C] ", "[E] ", "[W] ", "[I] ", "[D] "}

// 日志服务的实体类
type LibLogger struct {
	// 锁
	*sync.Mutex

	// 日志的等级
	level int

	// 日志的方法记录等级 -1为不记录
	funcCallDepth  int
	enableFuncCall bool

	// 是不是异步记录
	async bool

	// 同步策略
	wg sync.WaitGroup

	// 队列
	msgChan chan LoggerMsg

	// 信号
	signalChan chan int

	// 所有的注册的日志
	adapters map[string]Logger2

	// 默认的等级
	defaultLevel int
}

const defaultAsyncMsgLen = 1e3

// 初始化一个日志服务
func New() *LibLogger {
	l := new(LibLogger)
	l.Mutex = new(sync.Mutex)
	// 默认是同步
	l.async = false
	// 从最低等级开始记录
	l.level = LogLevelDebug
	l.adapters = make(map[string]Logger2)
	l.funcCallDepth = 3
	l.enableFuncCall = true
	l.signalChan = make(chan int, 1)
	l.defaultLevel = LogLevelInfo
	return l
}

// 注册一个日志的服务
func (l *LibLogger) Register(name string, config interface{}, log Logger2) error {
	l.Lock()
	defer l.Unlock()
	if log == nil {
		panic(fmt.Errorf("logs2 :register is log is nil [%s]", name))
	}

	// 如果已经注册，直接返回错误
	if _, adapater := l.adapters[name]; adapater {
		return fmt.Errorf("logs2 :register is log is registed [%s]", name)
	}
	// 初始化
	if err := log.Init(config); err != nil {
		return err
	}

	l.adapters[name] = log
	return nil
}

// 删除一个日志的服务
func (l *LibLogger) DelLogger(name string) error {
	l.Lock()
	defer l.Unlock()
	// 找到日志
	adapter := l.adapters[name]
	// 友善的清理
	adapter.Destroy()
	// 删除
	delete(l.adapters, name)
	return nil
}

// 异步日志处理
func (l *LibLogger) Async(worker, chanSize int) *LibLogger {

	l.Lock()
	defer l.Unlock()

	if worker <= 0 {
		panic(fmt.Errorf("logs2 : worker cannot be less than 1 but [%d]", worker))
	}
	if chanSize <= 1 {
		l.async = false
		return l
	}

	l.async = true
	l.msgChan = make(chan LoggerMsg, chanSize)
	l.wg.Add(1)
	defaultLogMsgPool = &sync.Pool{
		New: func() interface{} {
			return &DefaultLoggerMsg{}
		},
	}

	// 开始日志记录
	for i := 0; i < worker; i++ {
		go l.asyncLogger()
	}
	return l
}

func (l *LibLogger) write2Logger(when time.Time, msg string, level int) {
	for name, adapter := range l.adapters {
		err := adapter.WriteMsg(when, msg, level)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to write to [%s], error:[%v]", name, err)
		}
	}
}

func (l *LibLogger) writeMsg(logLevel int, msg string, v ...interface{}) error {
	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}

	// 小于这个等级的才可以打印
	if logLevel > l.level {
		return nil
	}

	when := time.Now()
	msg = l.formatMsg(when, msg, logLevel)

	if l.async {
		lm := defaultLogMsgPool.Get().(*DefaultLoggerMsg)
		lm.MsgLevel = logLevel
		lm.Message = msg
		lm.CreateAt = when
		l.msgChan <- lm
	} else {
		l.write2Logger(when, msg, logLevel)
	}

	return nil
}

func (l *LibLogger) formatMsg(when time.Time, msg string, level int) string {

	if l.enableFuncCall {
		_, file, line, ok := runtime.Caller(l.funcCallDepth)
		if !ok {
			file = "???"
			line = 0
		}
		_, filename := path.Split(file)

		msg = fmt.Sprintf("[%s:%d] %s", filename, line, msg)
	}

	if level <= LogLevelDebug {
		msg = levelPrefix[level-1] + msg
	}

	return msg
}

// 写消息
func (l *LibLogger) WriteMsg(msg LoggerMsg) error {
	if l.async {
		l.msgChan <- msg
	} else {
		str := l.formatMsg(msg.When(), msg.Msg(), msg.Level())
		l.write2Logger(msg.When(), str, msg.Level())
	}

	return nil
}

// writer interface
func (l *LibLogger) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	// 把回车删除
	if p[len(p)-1] == '\n' {
		p = p[0 : len(p)-1]
	}

	err = l.writeMsg(l.defaultLevel, string(p))

	if err == nil {
		return len(p), err
	}
	return 0, err
}

func (l *LibLogger) asyncLogger() {
	flag := false
	for {
		select {
		case msg := <-l.msgChan:
			l.write2Logger(msg.When(), msg.Msg(), msg.Level())
			defaultLogMsgPool.Put(msg)
		case sg := <-l.signalChan:
			l.flush()
			if sg == LoggerSignalClose {
				for _, adapter := range l.adapters {
					adapter.Destroy()
				}
				l.adapters = nil
				flag = true
			}
			l.wg.Done()
		}
		if flag {
			break
		}
	}
}

// flush msg
func (l *LibLogger) Flush() {
	if l.async {
		l.signalChan <- LoggerSignalFlush
		l.wg.Wait()
		l.wg.Add(1)
		return
	}
	l.flush()
}

func (l *LibLogger) Close() error {
	if l.async {
		l.signalChan <- LoggerSignalClose
		l.wg.Wait()
		l.wg.Add(1)
		return nil
	}
	l.flush()
	return nil
}

// 重置所有的日志
func (l *LibLogger) Reset() {
	l.Flush()
	for _, adapter := range l.adapters {
		adapter.Destroy()
	}
	l.adapters = nil
}

func (l *LibLogger) flush() {
	if l.async {
		for {
			if len(l.msgChan) > 0 {
				msg := <-l.msgChan
				l.write2Logger(msg.When(), msg.Msg(), msg.Level())
				defaultLogMsgPool.Put(msg)
				continue
			}
			break
		}
	}

	for _, adapter := range l.adapters {
		adapter.Flush()
	}
}

func (l *LibLogger) SetDefaultLevel(level int) *LibLogger {
	l.defaultLevel = level
	return l
}

// 设置，日志等级
func (l *LibLogger) SetLevel(val int) *LibLogger {
	l.level = val
	return l
}

// 设置呼叫的深度
func (l *LibLogger) SetFuncCallDepth(d int) *LibLogger {
	l.funcCallDepth = d
	return l
}

// 设置可以设置func
func (l *LibLogger) EnableFuncall(b bool) *LibLogger {
	l.enableFuncCall = b
	return l
}
