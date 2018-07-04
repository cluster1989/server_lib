package libdispatcher_test

import (
	"fmt"
	"testing"

	"github.com/wuqifei/server_lib/libdispatcher"
)

type JobTest struct {
	Num uint
}

func (j *JobTest) DoJobTask() error {
	fmt.Printf("i am doing my job now [%d]\n", j.Num)
	return nil
}

func TestDispatcher(t *testing.T) {

	dispatcher := libdispatcher.New(2, 10)
	for i := 0; i < 100; i++ {
		j := new(JobTest)
		j.Num = uint(i)
		dispatcher.Job.Enqueue(j)
	}
}
