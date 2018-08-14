package libdispatcher

type Dispatcher struct {
	WorkerPool chan chan Job
	Job        JobQueue
	maxWorkers uint
	workers    []*Worker
}

func New(worker, job uint) *Dispatcher {
	queue := InitJobQueue(job)
	dispatcher := NewDispatcher(worker)
	dispatcher.Job = queue
	dispatcher.workers = make([]*Worker, worker)
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
		d.workers = append(d.workers, worker)
	}

	go d.dispatch()
}

func (d *Dispatcher) Close() error {
	// 关闭
	d.Job.Close()
	close(d.WorkerPool)
	for _, w := range d.workers {
		w.Close()
	}
	return nil
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
