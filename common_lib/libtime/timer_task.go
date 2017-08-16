package libtime

import "time"

const (
	MIN_TIMER_INTERVAL = 1 * time.Millisecond // 循环定时器的最小时间间隔
)

type TimerTask struct {
	id          int64             //task 序号
	firetime    time.Time         //触发时间
	interval    time.Duration     //循环时间间隔
	count       int64             //执行次数
	alreadyExec int64             //已经执行次数
	timeout     *TimerTaskTimeOut //超时
	index       int               //在heap中的位置
}

func NewTask(interval time.Duration, count int64, to *TimerTaskTimeOut) *TimerTask {
	if interval < MIN_TIMER_INTERVAL {
		interval = MIN_TIMER_INTERVAL
	}
	task := &TimerTask{}
	task.id = timerIds.GetAndIncrement()
	task.firetime = time.Now().Add(interval)
	task.interval = interval
	task.count = count
	task.alreadyExec = 0
	task.timeout = to
	return task
}
