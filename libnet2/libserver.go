package libnet2

import (
	"io"
	"net"
	"strings"
	"time"

	"github.com/wuqifei/server_lib/concurrent"
)

// 默认的服务
type defaultLibServer struct {
	listener      net.Listener
	sessionOption *SessionOption2
	serverOption  *NetOption
	errorChan     chan error
	connCount     *concurrent.AtomicInt32
}

// 新建服务器
func newServer() *defaultLibServer {
	s := new(defaultLibServer)
	//  新建一个错误的通道
	s.errorChan = make(chan error, 10)
	s.connCount = concurrent.NewAtomicInt32(0)
	return s
}

func (s *defaultLibServer) Listener() net.Listener {
	return s.listener
}

// 关闭
func (s *defaultLibServer) Close() {

	s.listener.Close()
}

// 需要异步启动
func (s *defaultLibServer) Run() {

	for i := 0; i < s.serverOption.Workers; i++ {
		go s.run()
	}
}

func (s *defaultLibServer) run() {
	// 接收监听的连接
	var delay time.Duration
	for {
		// 接收监听

		if s.serverOption.MaxConn > 0 && s.serverOption.MaxConn < s.connCount.Get() {
			// 超过最大连接，等待
			continue
		}
		conn, err := s.listener.Accept()
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
				if ServerErrorBlock != nil {
					ServerErrorBlock(io.EOF)
				}

				return
			}

			if ServerErrorBlock != nil {
				ServerErrorBlock(err)
			}

			return
		}

		session := newDefaultSession(conn, s.sessionOption)

		if ServerSessionBlock != nil {
			ServerSessionBlock(session)
		}
		session.onClose = func(sess Session2Interface) {
			s.connCount.DecrementAndGet()
		}
		session.onError = SessionErrorBlock
		session.onRecv = SessionRecvBlock
		session.Accept()
		s.connCount.IncrementAndGet()
	}
}
