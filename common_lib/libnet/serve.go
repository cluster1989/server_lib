package libnet

import (
	"net"
	"time"

	"github.com/wqf/common_lib/libio"
	"github.com/wqf/common_lib/libnet/def"
	"github.com/wqf/common_lib/libnet/session"
)

type ServerOptions struct {
	// 类型
	Network string

	// 地址
	Address string

	// cpu大小头设置
	IsLittleIndian bool

	// 接收和发送的队列大小
	SendQueueBuf int
	RecvQueueBuf int

	//接收和发送的超时时间
	SendTimeOut time.Duration
	RecvTimeOut time.Duration

	//server 心跳时间
	HeartBeatTime time.Duration

	// 允许超时次数
	ReadTimeOutTimes int
}

func Serve(options *ServerOptions) *Server {
	listener, err := net.Listen(options.Network, options.Address)
	if err != nil {
		panic("server init error")
	}
	proto := libio.New(options.IsLittleIndian)
	server := NewServer(listener, proto)
	server.Options = options
	return server
}

// 测试用
func Connect(network, address string, p def.Protocol) (*session.Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return session.NewSession(p.NewCodec(conn), 0, 0, 0, 0, 0), nil
}

// 测试用
func ConnectTimeout(network, address string, timeout time.Duration, p def.Protocol) (*session.Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	return session.NewSession(p.NewCodec(conn), 0, 0, 0, 0, 0), nil
}
