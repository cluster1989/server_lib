package libnet

import (
	"net"
	"time"
)

func Serve(network, address string, p Protocol, sendChanSize int) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, p, sendChanSize), nil
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
