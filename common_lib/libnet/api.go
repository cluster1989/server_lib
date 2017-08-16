package libnet

import (
	"net"
	"time"

	"github.com/wqf/common_lib/libnet/libio"
	"github.com/wqf/zone_server/zone_client"
)

func NewServerAndRun(network, address string, isLittleIndian bool) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	proto := libio.New(isLittleIndian)
	server := NewServer(listener, p, 0)
	go runLoop(server)
}

func runLoop(server *Server) {
	for {
		session, err := s.Server.Accept() //接收客户端连接
		if err != nil {
			//记录错误
			continue
		}

		//session 处理
		client := zone_client.New(session) //因为库已经帮助持有了对象
		go client.ClientLoop()
	}
}

func Connect(network, address string, p Protocol, sendChanSize int) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewSession(p.NewCodec(conn), sendChanSize), nil
}

func ConnectTimeout(network, address string, timeout time.Duration, protocol Protocol, sendChanSize int) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	return NewSession(protocol.NewCodec(conn), sendChanSize), nil
}
