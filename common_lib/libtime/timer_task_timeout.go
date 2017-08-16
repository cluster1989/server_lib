package libtime

type TimerTaskTimeOut struct {
	Callback func(val interface{}) //回调
	Content  interface{}           //希望带过去的参数
}

func NewTimerTaskTimeOut(content interface{}, callback func(val interface{})) *TimerTaskTimeOut {
	t := &TimerTaskTimeOut{}
	t.Content = content
	t.Callback = callback
	return t
}
