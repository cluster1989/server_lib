package message

import (
	"fmt"

	"github.com/wqf/common_lib/libnet/handler"
)

var (
	Routes map[uint16]handler.MessageHandlerWithRet
)

func init() {
	Routes = make(map[uint16]handler.MessageHandlerWithRet)
}

// 普通消息
func Register(msgType uint16, handler handler.MessageHandlerWithRet) {
	if _, ok := Routes[msgType]; ok {
		panic(fmt.Sprintf("trying to register message %d twice", msgType))
	}
	Routes[msgType] = handler
}

func GetHandler(msgType) handler.MessageHandlerWithRet {
	handler, ok := Routes[msgType]
	if !ok {
		return nil
	}
	return handler
}
