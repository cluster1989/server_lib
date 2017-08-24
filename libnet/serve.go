package libnet

import (
	"net"
	"time"

	"github.com/wuqifei/server_lib/libio"
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

	// 最大的接收字节数
	MaxRecvBufferSize int

	// 最大的发送字节数
	MaxSendBufferSize int
}

func Serve(options *ServerOptions) *Server {
	listener, err := net.Listen(options.Network, options.Address)
	if err != nil {
		panic("server init error")
	}
	proto := libio.New(options.IsLittleIndian, options.MaxRecvBufferSize, options.MaxSendBufferSize)
	server := NewServer(listener, proto)
	server.Options = options
	return server
}
