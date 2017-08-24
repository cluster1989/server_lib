package libtime

import "github.com/wuqifei/server_lib/concurrent"

//为每个定时器生成id
var timerIds *concurrent.AtomicInt64

func init() {
	//给定初始数值
	timerIds = concurrent.NewAtomicInt64(1)
}
