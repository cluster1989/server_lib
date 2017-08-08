package libtime

import "testing"
import "time"

var gtask *LTicker

func TestTicker(t *testing.T) {
	task, _ := NewTicker("abc", time.Duration(1)*time.Second, true, repeat)
	gtask = task
	task.Start()
	time.Sleep(time.Duration(5) * time.Second)
	task.Pause()

	time.Sleep(time.Duration(2) * time.Second)
	task.Resume()

	time.Sleep(time.Duration(5) * time.Second)
}

func TestOneTimeTicker(t *testing.T) {
	task, _ := NewTicker("abc", time.Duration(1)*time.Second, false, repeat)
	gtask = task
	task.Start()
	time.Sleep(time.Duration(2) * time.Second)
	task.Close()
}

func repeat() {
	print("aciton:", time.Now().Second(), "count:", gtask.Count, "\n")
}
