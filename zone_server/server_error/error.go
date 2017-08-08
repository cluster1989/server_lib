package server_error

import (
	"github.com/golang/protobuf/proto"
	"github.com/wqf/zone_server/pb"
)

const (
	// 登陆错误
	LoginErr = iota
	// 注册错误
	RegErr
)

// 错误处理
func NewError(code int32) *pb.RawMsg {
	msg := new(pb.RawMsg)
	errMsg := new(pb.ErrorAck)
	errMsg.Errid = proto.Int32(code)
	msg.MsgData = msg
	msg.MsgId = 999 //错误码
	return msg
}
