package message

import (
	"fmt"

	"github.com/wuqifei/server_lib/libnet/def"
)

var (
	heartBeatType uint16
)

// 心跳信息
func RegisterHeartBeat(msgType uint16, handler def.MessageHandlerWithRet) {
	if _, ok := Routes[msgType]; ok {
		panic(fmt.Sprintf("trying to register message %d twice", msgType))
	}
	heartBeatType = msgType
	Routes[msgType] = handler
}

func GetHeartBeatHandler() def.MessageHandlerWithRet {
	return GetHandler(heartBeatType)
}
