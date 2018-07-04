package libdispatcher

type Dispatcher struct {
	WorkerPool chan chan Job
	Job        JobQueue
	maxWorkers uint
}

func New(worker, job uint) *Dispatcher {
	queue := InitJobQueue(job)
	dispatcher := NewDispatcher(worker)
	dispatcher.Job = queue
	dispatcher.Run()
	return dispatcher
}

func NewDispatcher(maxWorkers uint) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{WorkerPool: pool, maxWorkers: maxWorkers}
}

func (d *Dispatcher) Run() {
	for i := 0; i < int(d.maxWorkers); i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		// 获取job
		case job := <-d.Job:
			go func(job Job) {

				jobChannel := <-d.WorkerPool
				jobChannel <- job
			}(job)
		}
	}
}
