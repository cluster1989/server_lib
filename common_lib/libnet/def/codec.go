package def

type Codec interface {
	Receive() ([]byte, error)
	Send(interface{}) error
	Close() error
}

type MessageCodec interface {

	// 序列化
	MessageSerialize() ([]byte, error)

	// 消息类型
	MessageType() int32
}
