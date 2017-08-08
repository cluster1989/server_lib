package zone_route

import (
	"github.com/golang/protobuf/proto"
	"github.com/wqf/common_lib/codec"
	"github.com/wqf/zone_server/handler"
	"github.com/wqf/zone_server/pb"
)

var Routes map[uint16]handler.MessageHandlerWithRet
var ExceptionRouteIds = [...]uint16{10001, 10002, 10003}

func init() {
	Routes = make(map[uint16]handler.MessageHandlerWithRet)
}

func RegisterAllRoute() {
	Routes[10000] = Login
}

func PackData(data *pb.RawMsg) (b []byte, err error) {

	pbPacket, pberr := proto.Marshal(data.MsgData.(proto.Message))
	if pberr != nil {
		err = pberr
		return
	}
	packet := make([]byte, 2)
	codec.PutUint16BE(packet, data.MsgId)
	packet = append(packet, pbPacket...)
	b = packet
	return
}
