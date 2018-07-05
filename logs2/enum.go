package logs2

// 消息的等级
const (
	// 紧急
	LogLevelEmergency = iota
	// 严格的
	LogLevelCritical
	// 错误的
	LogLevelError
	// 警告
	LogLevelWarning
	// 一般信息
	LogLevelInfo
	// 测试信息
	LogLevelDebug
)

const (
	// 刷新日志
	LoggerSignalFlush = iota + 1
	// 关闭日志
	LoggerSignalClose
)
