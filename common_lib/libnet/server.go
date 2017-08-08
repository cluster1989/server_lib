package libnet

import (
	"io"
	"net"
	"strings"
	"time"
)

type Server struct {
	manager      *Manager
	listerer     net.Listener
	protocol     Protocol
	sendChanSize int
}

func NewServer(l net.Listener, p Protocol, sendChanSize int) *Server {
	return &Server{
		manager:      NewManager(),
		listerer:     l,
		protocol:     p,
		sendChanSize: sendChanSize,
	}
}

func (server *Server) Listener() net.Listener {
	return server.listerer
}

func (server *Server) Accept() (*Session, error) {
	var delay time.Duration
	for {
		conn, err := server.listerer.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if delay == 0 {
					delay = 5 * time.Millisecond
				} else {
					delay *= 2
				}
				if max := 1 * time.Second; delay > max {
					delay = max
				}
				time.Sleep(delay)
				continue
			}
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil, io.EOF
			}
			return nil, err
		}

		return server.manager.NewSession(server.protocol.NewCodec(conn), server.sendChanSize), nil
	}
}

func (s *Server) Stop() {
	s.listerer.Close()
	s.manager.Dispose()
}
