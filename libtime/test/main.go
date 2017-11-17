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
	task1 := libtime.NewTimerTaskTimeOut("测试任务1", func(val interface{}) {
		fmt.Printf("测试任务1111 callback: %v;timeis:%s \n", val, time.Now().Format("2006-01-02 15:04:05"))
	})

	task2 := libtime.NewTimerTaskTimeOut("测试任务2", func(val interface{}) {
		fmt.Printf("测试任务2222 callback: %v;timeis:%s \n", val, time.Now().Format("2006-01-02 15:04:05"))
	})

	timerId1 := wheel.AddTask(time.Duration(1)*time.Second, -1, task1)
	timerId2 := wheel.AddTask(time.Duration(5)*time.Second, -1, task2)

	go delTimer(timerId1)
	go delTimer(timerId2)
	signal.InitSignal()
}

func delTimer(timerId int64) {
	return
	//先过个5秒再删除
	time.Sleep(time.Duration(timerId) * time.Second)
	fmt.Printf("删除任务：%d\n", timerId)
	wheel.CancelTimer(timerId)
	for i := 0; i < 5; i++ {

		go addNewTask()
	}
}

func addNewTask() {
	task3 := libtime.NewTimerTaskTimeOut("测试任务不知道多少了", func(val interface{}) {
		fmt.Printf("测试任务tttt callback: %v;timeis:%s \n", val, time.Now().Format("2006-01-02 15:04:05"))
	})
	timerId := wheel.AddTask(time.Duration(3)*time.Second, -1, task3)

	go delTimer(timerId)
}
