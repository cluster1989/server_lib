package def

type MessageHandlerWithRet func(content []byte) (msg MessageCodec, err error)
