package libtime

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

const (
	bufferSize               = 1024
	tickPeriod time.Duration = 500 * time.Millisecond
)

type TimerWheel struct {
	timeOutChan chan *TimerTaskTimeOut
	timers      *TimerHeap
	ticker      *time.Ticker
	waitGroup   sync.WaitGroup
	addChan     chan *TimerTask
	cancelChan  chan int64
	stopchan    chan bool
	sizeChan    chan int
}

func NewTimerWheel() *TimerWheel {
	wheel := &TimerWheel{}
	wheel.timeOutChan = make(chan *TimerTaskTimeOut, bufferSize)
	wheel.timers = NewHeep()
	wheel.ticker = time.NewTicker(tickPeriod)
	wheel.addChan = make(chan *TimerTask, bufferSize)
	wheel.cancelChan = make(chan int64, bufferSize)
	wheel.sizeChan = make(chan int)
	heap.Init(wheel.timers)
	go func() {
		wheel.start()
	}()
	return wheel
}

func (w *TimerWheel) AddTask(interval time.Duration, count int64, to *TimerTaskTimeOut) int64 {
	if to == nil {
		return -1
	}
	task := NewTask(interval, count, to)
	w.addChan <- task
	w.waitGroup.Add(1)
	return task.id
}

func (w *TimerWheel) Size() int {
	return <-w.sizeChan
}

// 关闭timer，关闭时候，使用这个方法
func (w *TimerWheel) CancelTimer(id int64) {
	w.cancelChan <- id
}

func (w *TimerWheel) Stop() {
	//执行，并清空所有task
	w.stopchan <- true
	//等待所有的都执行完毕
	w.waitGroup.Wait()
	w.ticker.Stop()
	close(w.timeOutChan)
	close(w.cancelChan)
	close(w.addChan)
	close(w.stopchan)
	close(w.sizeChan)
}

func (w *TimerWheel) getLattestTimer() []*TimerTask {
	expired := make([]*TimerTask, 0)
	for w.timers.Len() > 0 {
		task := heap.Pop(w.timers).(*TimerTask)
		nextFireTime := task.firetime
		elasped := time.Since(nextFireTime).Seconds()
		if elasped > 1.0 {
			fmt.Printf("libtime:timer exec error with 1 second not exec\n")
		}
		if elasped > 0.0 {
			//时间未到
			expired = append(expired, task)
			continue
		} else {
			heap.Push(w.timers, task)
			break
		}
	}
	return expired
}

func (w *TimerWheel) Remove(id int64) {
	if id <= 0 {
		return
	}
	index := w.timers.GetIndexByID(id)
	if index >= 0 {
		heap.Remove(w.timers, index)
	}
	w.waitGroup.Done()
}

func (w *TimerWheel) Flush() {
	for w.timers.Len() > 0 {
		task := heap.Pop(w.timers).(*TimerTask)
		//执行调用
		task.timeout.Callback(task.timeout.Content)
		w.waitGroup.Done()
	}
}

func (w *TimerWheel) updateTimers(timers []*TimerTask) {
	if timers == nil {
		return
	}

	for _, t := range timers {
		if t.count < 0 || t.alreadyExec < (t.count-1) {
			t.firetime = t.firetime.Add(t.interval)
			if time.Since(t.firetime).Seconds() >= 1.0 {
				t.firetime = time.Now()
			}
			heap.Push(w.timers, t)
		} else {
			// 这里时真正删除
			//在这里将计数器减去1
			w.waitGroup.Done()
		}
	}
}

func (w *TimerWheel) start() {
	for {
		select {
		case id := <-w.cancelChan:
			//取消某个
			w.Remove(id)
		case w.sizeChan <- w.timers.Len():
		case <-w.stopchan:
			w.Flush() //flush 之后要直接return 破开循环
			return
		case task := <-w.addChan:
			heap.Push(w.timers, task)
		case timeoutTask := <-w.timeOutChan:
			timeoutTask.Callback(timeoutTask.Content)
		case <-w.ticker.C:
			timers := w.getLattestTimer()
			for _, t := range timers {
				t.alreadyExec++
				//执行一次
				w.timeOutChan <- t.timeout
			}
			w.updateTimers(timers)
		}
	}
}
