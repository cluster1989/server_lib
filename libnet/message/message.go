package message

import (
	"fmt"

	"github.com/wuqifei/server_lib/libnet/def"
)

var (
	Routes map[uint16]def.MessageHandlerWithRet
)

func init() {
	Routes = make(map[uint16]def.MessageHandlerWithRet)
}

// 普通消息
func Register(msgType uint16, def def.MessageHandlerWithRet) {
	if _, ok := Routes[msgType]; ok {
		panic(fmt.Sprintf("trying to register message %d twice", msgType))
	}
	Routes[msgType] = def
}

func GetHandler(msgType uint16) def.MessageHandlerWithRet {
	def, ok := Routes[msgType]
	if !ok {
		return nil
	}
	return def
}
