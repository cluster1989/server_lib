package libtime

type TimerHeap struct {
	timers []*TimerTask
}

func NewHeep() *TimerHeap {
	heap := &TimerHeap{}
	heap.timers = make([]*TimerTask, 0)
	return heap
}

func (heap *TimerHeap) GetIndexByID(id int64) int {
	for _, item := range heap.timers {
		if item.id == id {
			return item.index
		}
	}
	return -1
}

func (heap *TimerHeap) Len() int {
	return len(heap.timers)
}

func (heap *TimerHeap) Less(i, j int) bool {
	t1, t2 := heap.timers[i].firetime, heap.timers[j].firetime
	if t1.Before(t2) {
		return true
	}
	if t2.Before(t1) {
		return false
	}

	return heap.timers[i].index < heap.timers[j].index
}

func (heap *TimerHeap) Swap(i, j int) {
	var tmp *TimerTask
	tmp = heap.timers[i]
	heap.timers[i] = heap.timers[j]
	heap.timers[j] = tmp
	heap.timers[i].index = i
	heap.timers[j].index = j
}

func (heap *TimerHeap) Push(task interface{}) {
	n := heap.Len()
	timer := task.(*TimerTask)
	timer.index = n
	heap.timers = append(heap.timers, timer)
}

func (heap *TimerHeap) Pop() interface{} {
	n := heap.Len()
	task := heap.timers[n-1]
	task.index = -1
	heap.timers = heap.timers[0 : n-1]
	return task
}
