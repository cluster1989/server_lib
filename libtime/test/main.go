package main

import (
	"fmt"
	"time"

	"github.com/wuqifei/server_lib/libtime"
	"github.com/wuqifei/server_lib/signal"
)

var (
	wheel *libtime.TimerWheel
)

func main() {

	wheel = libtime.NewTimerWheel()
	task1 := libtime.NewTimerTaskTimeOut("linik1", func(val interface{}) {
		fmt.Printf("recv task1 callback: %v\n", val)

		fmt.Printf("timecount  :%d \n", time.Now().Second())
	})

	task2 := libtime.NewTimerTaskTimeOut("linik2", func(val interface{}) {
		fmt.Printf("recv task2 callback: %v\n", val)

		fmt.Printf("timecount  :%d \n", time.Now().Second())
	})

	timerId1 := wheel.AddTask(time.Duration(1)*time.Second, -1, task1)
	timerId2 := wheel.AddTask(time.Duration(2)*time.Second, -1, task2)

	go delTimer1(timerId1)
	go delTimer2(timerId2)
	signal.InitSignal()
}

func addNewTask() {
	task3 := libtime.NewTimerTaskTimeOut("linik3", func(val interface{}) {
		fmt.Printf("recv task3 callback: %v\n", val)

		fmt.Printf("timecount  :%d \n", time.Now().Second())
	})
	timerId3 := wheel.AddTask(time.Duration(3)*time.Second, -1, task3)

	go delTimer1(timerId3)
}

func delTimer1(timerId1 int64) {
	time.Sleep(time.Duration(5) * time.Second)
	wheel.CancelTimer(timerId1)
	go addNewTask()
	go addNewTask()
}

func delTimer2(timerId2 int64) {
	time.Sleep(time.Duration(20) * time.Second)
	wheel.CancelTimer(timerId2)
}
