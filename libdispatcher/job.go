package libdispatcher

type Job interface {
	DoJobTask() error
}

type JobQueue chan Job

func InitJobQueue(job uint) JobQueue {
	return make(chan Job, job)
}

func (j JobQueue) Enqueue(job Job) {
	j <- job
}

func (j JobQueue) EmptyQueue(job Job) {
	j.Close()
	j = make(chan Job)
}

func (j JobQueue) Close() error {
	close(j)
	return nil
}
