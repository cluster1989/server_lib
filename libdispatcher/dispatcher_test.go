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

func BenckmarkDispatcher(b *testing.B) {
	dispatcher := libdispatcher.New(4, 10)
	for i := 0; i < b.N; i++ {
		j := &JobTest{
			Num: uint(i),
		}
		dispatcher.Job.Enqueue(j)
	}
}
