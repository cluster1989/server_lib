package message

import (
	"fmt"
	"time"

	"github.com/wqf/common_lib/libnet/def"
)

var (
	heartBeatType     uint16
	heartBeatDuration time.Duration
)

// 心跳信息
func RegisterHeartBeat(msgType uint16, duration time.Duration, handler def.MessageHandlerWithRet) {
	heartBeatDuration = duration
	if _, ok := Routes[msgType]; ok {
		panic(fmt.Sprintf("trying to register message %d twice", msgType))
	}
	heartBeatType = msgType
	Routes[msgType] = handler
}

func GetHeartBeatHandler() def.MessageHandlerWithRet {
	return GetHandler(heartBeatType)
}

func GetHeartBeatDuration() time.Duration {
	return heartBeatDuration
}
