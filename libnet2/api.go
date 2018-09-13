package libnet2

import (
	"net"
	"time"
)

var (
	// 服务收到错误的回调
	ServerErrorBlock OnError

	// 服务收到 session的回调
	ServerSessionBlock OnSession

	// session收到信息
	SessionRecvBlock OnSessRecv
	// session关闭
	SessionCloseBlock OnSessClose
	// session错误
	SessionErrorBlock OnSessError

	// 解析的对象，必须要有
	ServerPacket PacketInterface
)

// 创建一个默认的服务
func New() (LibserverInterface, error) {
	option := DefaultOption()
	sessionOption := DefaultSessionOption()
	return NewWithOption(option, sessionOption)
}

// 用配置新建一个服务
func NewWithOption(option *NetOption, sessionOption *SessionOption2) (LibserverInterface, error) {

	listener, err := net.Listen(option.Network, option.Address)

	if err != nil {
		return nil, err
	}
	server := newServer()
	server.listener = listener
	server.serverOption = option
	server.sessionOption = sessionOption
	return server, nil
}

// 默认的session服务
func DefaultSessionOption() *SessionOption2 {
	option := new(SessionOption2)
	// 60s
	option.ReadTimeout = time.Second * time.Duration(60)
	// 允许3次超时
	option.ReadTimeoutTimes = 3
	// 发送队列缓冲条数
	option.RecvChanSize = 10
	// 接收队列缓冲条数
	option.SendChanSize = 10

	return option
}

// 新建默认的配置
func DefaultOption() *NetOption {
	option := new(NetOption)
	// 默认端口10001
	option.Address = ":10001"
	option.MaxConn = -1
	option.Network = "tcp"
	option.Workers = 4
	return option
}

// 默认的服务session
func DefaultSession(conn net.Conn, sessionOption *SessionOption2) Session2Interface {
	return newDefaultSession(conn, sessionOption)
}
