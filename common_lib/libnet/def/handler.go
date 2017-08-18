package def

// 传入接收到的信息，返回函数处理完之后的信息
type MessageHandlerWithRet func(content []byte) (ack MessageAckCodec, err error)
