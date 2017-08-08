package libtime

import (
	"errors"
	"runtime"
	"sync"
	"time"
)

const (
	maxInt64         = (1<<63 - 1)
	infiniteDuration = time.Duration(maxInt64)
)

type LTickerHandler func()

type LTicker struct {
	sync.Mutex
	Name         string
	Status       int
	IntervalTime time.Duration //间隔时间
	Count        uint64        // 表示已执行的次数
	CreateTime   time.Time
	statusChan   chan int
	Repeat       bool           //是否需要循环执行
	Handler      LTickerHandler //回调
}

func NewTicker(name string, intervalTime time.Duration, Repeat bool, handle LTickerHandler) (task *LTicker, err error) {
	if len(name) == 0 {
		task = nil
		err = errors.New("task名字不能为空")
		return
	}
	t := &LTicker{
		Name:         name,
		IntervalTime: intervalTime,
		CreateTime:   time.Now(),
		statusChan:   make(chan int),
		Count:        0,
		Repeat:       Repeat,
		Handler:      handle,
	}
	task = t
	err = nil
	return
}

func (t *LTicker) Run() {
	timer := time.NewTicker(t.IntervalTime)
	for {
		select {
		case <-timer.C:
			if t.Status == Pause {
				runtime.Gosched() //让出时间片
				continue
			}

			t.Handler()
			t.Count += 1
			if t.Repeat == false {
				t.Stop()
			}
		case status := <-t.statusChan:
			switch status {
			case Stop:
				timer.Stop()
				return
			case Pause:
			case Resume:
			}
		}
	}
}

func (t *LTicker) Start() {
	t.Status = Running
	go t.Run()
}

func (t *LTicker) Stop() {
	if t.Status == Stop {
		return
	}
	t.statusChan <- Stop
	t.Status = Stop
}

func (t *LTicker) Pause() {
	if t.Status != Running {
		return
	}
	t.statusChan <- Pause
	t.Status = Pause
}

func (t *LTicker) Resume() {
	if t.Status != Pause {
		return
	}
	t.statusChan <- Resume
	t.Status = Resume
}

func (t *LTicker) Close() {
	close(t.statusChan)
	t.Handler = nil
}
