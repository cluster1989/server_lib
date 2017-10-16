package libnet

import (
	"net"

	"github.com/wuqifei/server_lib/libio"
	"github.com/wuqifei/server_lib/libsession"
	"github.com/wuqifei/server_lib/logs"
)

type ServerOptions struct {
	// 类型
	Network string

	// 地址
	Address string

	SessionOption libsession.Options

	// 最大的接收字节数
	MaxRecvBufferSize int

	// 最大的发送字节数
	MaxSendBufferSize int
}

func Serve(options *ServerOptions) *Server {
	listener, err := net.Listen(options.Network, options.Address)
	if err != nil {
		panic(err)
	}
	proto := libio.New(options.SessionOption.IsLittleEndian, options.MaxRecvBufferSize, options.MaxSendBufferSize)
	server := NewServer(listener, proto)
	server.Options = options
	logs.Informational("libnet:Server Start")
	//打印输出的logger
	logs.Informational("libnet:network(%s),listento(%s),", options.Network, options.Address)
	return server
}
