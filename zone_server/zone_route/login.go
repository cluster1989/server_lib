package zone_route

import (
	"github.com/golang/protobuf/proto"
	"github.com/wqf/zone_server/pb"
	"github.com/wqf/zone_server/server_error"
)

func Login(data ...interface{}) interface{} {
	if len(data) != 1 {
		packet, _ := PackData(server_error.NewError(server_error.LoginErr))
		return packet
	}
	dataByte := data[0].([]byte)
	loginPBReq := new(pb.TouristsLoginReq)
	if err := proto.Unmarshal(dataByte, loginPBReq); err != nil {
		packet, _ := PackData(server_error.NewError(server_error.LoginErr))
		return packet
	}
	//检测是否在clientmanager 里面有这个人了
	if *loginPBReq.Touristscot != "wqf" || *loginPBReq.Touristspwd != "123456" {

		packet, _ := PackData(server_error.NewError(server_error.LoginErr))
		return packet
	}

	ret := new(pb.RawMsg)
	ret.MsgId = 10001
	logAck := new(pb.TouristsLoginAck)
	strBack := "succeed"
	logAck.Str = &strBack
	ret.MsgData = logAck
	packet, _ := PackData(ret)
	return packet
}
