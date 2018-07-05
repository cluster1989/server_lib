package logs2

import (
	"fmt"
	"strings"
)

var libLogger *LibLogger

// 默认的日志
func DefaultLogger() *LibLogger {
	if libLogger == nil {
		libLogger = New()
	}
	return libLogger
}

// 紧急
func Emergency(f interface{}, v ...interface{}) {
	DefaultLogger().Emergency(formatLog(f, v...))
}

// 严格的
func Critical(f interface{}, v ...interface{}) {
	DefaultLogger().Critical(formatLog(f, v...))
}

// 错误的
func Error(f interface{}, v ...interface{}) {
	DefaultLogger().Error(formatLog(f, v...))
}

// 警告
func Warning(f interface{}, v ...interface{}) {
	DefaultLogger().Warning(formatLog(f, v...))
}

// 一般信息
func Info(f interface{}, v ...interface{}) {
	DefaultLogger().Info(formatLog(f, v...))
}

// 一般信息
func Debug(f interface{}, v ...interface{}) {
	DefaultLogger().Debug(formatLog(f, v...))
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
			//f interface{}
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
