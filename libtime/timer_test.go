package libtime

import (
	"fmt"
	"testing"
	"time"

	"github.com/wuqifei/server_lib/signal"
)

func TestTimer(t *testing.T) {
	wheel := NewTimerWheel()
	timeOutTask := NewTimerTaskTimeOut("haha", func(val interface{}) {
		fmt.Printf("recved task callback :%v \n", val)
		fmt.Printf("timecount  :%d \n", time.Now().Second())
	})
	timerId := wheel.AddTask(time.Duration(2)*time.Second, 10, timeOutTask)
	fmt.Printf("timer id:%d \n", timerId)

	signal.InitSignal()
}
