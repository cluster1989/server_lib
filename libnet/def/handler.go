package def

// 传入接收到的信息，返回函数处理完之后的信息
// 返回msgid，msg的byte ，如果有登陆注册则返回多一个uniqueid
// 会向上层发送，由本地注册的sessionID，由上层统一管理用户以及SessionID的对应关系
type MessageHandlerWithRet func(content []byte, sessionID uint64) (args []interface{})
