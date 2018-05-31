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
	clientGroup   *concurrent.ConcurrentIDGroupMap
}

// 新建服务器
func newServer() *defaultLibServer {
	s := new(defaultLibServer)
	//  新建一个错误的通道
	s.errorChan = make(chan error, 10)
	s.clientGroup = concurrent.NewCocurrentIDGroup()
	return s
}

func (s *defaultLibServer) Listener() net.Listener {
	return s.listener
}

// 删除某个session
func (s *defaultLibServer) DelSession(sessID uint64) error {
	sessInterface := s.clientGroup.Get(sessID)
	if sessInterface == nil {
		return nil
	}
	sess := sessInterface.(Session2Interface)
	err := sess.Close()
	s.clientGroup.Del(sessID)

	SessionCloseBlock(sess)
	return err
}

// 得到某个session
func (s *defaultLibServer) GetSession(sessID uint64) Session2Interface {
	sessInterface := s.clientGroup.Get(sessID)
	if sessInterface == nil {
		return nil
	}
	sess := sessInterface.(Session2Interface)

	return sess
}

// 设置session
func (s *defaultLibServer) SetSession(sess Session2Interface) {

	s.clientGroup.Set(sess.GetUniqueID(), sess)
}

// 得到全部session
func (s *defaultLibServer) GetAllSession() *concurrent.ConcurrentIDGroupMap {
	return s.clientGroup
}

// 关闭
func (s *defaultLibServer) Close() {

	s.listener.Close()
	//释放所有的连接
	s.clientGroup.Dispose()
}

// 需要异步启动
func (s *defaultLibServer) Run(procs int) {

	for i := 0; i < procs; i++ {
		go s.run()
	}
}

func (s *defaultLibServer) run() {
	// 接收监听的连接
	var delay time.Duration
	for {
		// 接收监听

		if s.serverOption.MaxConn > 0 && s.serverOption.MaxConn < s.clientGroup.Count() {
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
			s.DelSession(sess.GetUniqueID())
		}
		session.onError = SessionErrorBlock
		session.onRecv = SessionRecvBlock
		session.Accept()
		s.SetSession(session)
	}
}

// 刷新sessionid
func (s *defaultLibServer) UpdateSessionID(oldID, newID uint64) bool {

	oldSess := s.GetSession(oldID)
	if oldSess == nil {
		return false
	}
	newSess := s.GetSession(newID)
	if newSess != nil {
		return false
	}
	s.clientGroup.Del(oldID)

	sess := oldSess.(*defaultSession)
	sess.setUniqueID(newID)
	s.SetSession(sess)
	return true
}
