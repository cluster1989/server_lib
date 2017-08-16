package message

import (
	"fmt"
	"time"

	"github.com/wqf/common_lib/libnet/handler"
)

var (
	heartBeatType     HeartMessage
	heartBeatDuration time.Duration
)

// 心跳信息
func RegisterHeartBeat(msgType uint16, duration time.Duration, handler handler.MessageHandlerWithRet) {
	heartBeatDuration = duration
	if _, ok := Routes[msgType]; ok {
		panic(fmt.Sprintf("trying to register message %d twice", msgType))
	}
	heartBeatType = msgType
	Routes[msgType] = handler
}

func GetHeartBeatHandler() handler.MessageHandlerWithRet {
	return GetHandler(heartBeatType)
}

func GetHeartBeatDuration() time.Duration {
	return heartBeatDuration
}
