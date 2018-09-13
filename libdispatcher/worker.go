package libdispatcher

import (
	"fmt"
	"sync"

	"github.com/wuqifei/server_lib/concurrent"
)

type Worker struct {
	WorkerPool  chan chan Job
	JobChannel  chan Job
	quit        chan bool
	disposeOnce sync.Once
	closeFlag   *concurrent.AtomicBoolean
}

func NewWorker(workerPool chan chan Job) *Worker {
	return &Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
		closeFlag:  concurrent.NewAtomicBoolean(false),
	}
}

// 开始任务
func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				// 开始任务
				err := job.DoJobTask()
				if err != nil {
					fmt.Printf("job  task went wrong [%v]\n", err)
				}
			case <-w.quit:
				w.close()
				return
			}
		}
	}()
}

func (w *Worker) close() {
	w.closeFlag.Set(true)
	w.disposeOnce.Do(func() {
		close(w.quit)
		close(w.WorkerPool)
		close(w.JobChannel)
	})
}

func (w *Worker) Close() error {
	if w.closeFlag.Get() {
		w.quit <- true
	}
	return nil
}
