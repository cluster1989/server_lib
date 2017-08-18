package def

import "io"

type Protocol interface {
	NewCodec(rw io.ReadWriter) Codec
}

type Codec interface {
	Receive() ([]byte, error)
	Send(interface{}) error
	Close() error
}

type MessageAckCodec interface {

	// 这里的返回参数，当参数超过1个的时候，默认会有一个user unique id，来作为返回参数
	MessageSerialize() []interface{}

	// 消息类型
	MessageType() int32
}
