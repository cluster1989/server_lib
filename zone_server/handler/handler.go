package handler

// type MessageHandlerWithOutRet func(data ...interface{})
type MessageHandlerWithRet func(data ...interface{}) interface{}
