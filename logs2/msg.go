package logs2

import (
	"time"
)

// 日志的消息服务
type LoggerMsg interface {
	// 是什么等级
	Level() int
	// 消息体
	Msg() string
	// 发送时间
	When() time.Time
}

// 默认的消息对象
type DefaultLoggerMsg struct {
	MsgLevel int
	Message  string
	CreateAt time.Time
}

func (d *DefaultLoggerMsg) Level() int {
	return d.MsgLevel
}

func (d *DefaultLoggerMsg) Msg() string {
	return d.Message
}

func (d *DefaultLoggerMsg) When() time.Time {
	return d.CreateAt
}

// 紧急
func (l *LibLogger) Emergency(format string, v ...interface{}) {
	if LogLevelEmergency > l.level {
		return
	}
	l.writeMsg(LogLevelEmergency, format, v...)
}

// 严格的
func (l *LibLogger) Critical(format string, v ...interface{}) {
	if LogLevelCritical > l.level {
		return
	}
	l.writeMsg(LogLevelCritical, format, v...)
}

// 错误的
func (l *LibLogger) Error(format string, v ...interface{}) {
	if LogLevelError > l.level {
		return
	}
	l.writeMsg(LogLevelError, format, v...)
}

// 警告
func (l *LibLogger) Warning(format string, v ...interface{}) {
	if LogLevelWarning > l.level {
		return
	}
	l.writeMsg(LogLevelWarning, format, v...)
}

// 一般信息
func (l *LibLogger) Info(format string, v ...interface{}) {
	if LogLevelInfo > l.level {
		return
	}
	l.writeMsg(LogLevelInfo, format, v...)
}

// 一般信息
func (l *LibLogger) Debug(format string, v ...interface{}) {
	if LogLevelDebug > l.level {
		return
	}
	l.writeMsg(LogLevelDebug, format, v...)
}
