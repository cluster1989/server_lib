package def

// 传入接收到的信息，返回函数处理完之后的信息
// 返回msgid，msg的byte ，如果有登陆注册则返回多一个uniqueid
// 会向上层发送，这个消息是不是个野消息，因为会是游客发送的，由上层统一管理用户
type MessageHandlerWithRet func(content []byte, wildMsg bool) (args []interface{})
