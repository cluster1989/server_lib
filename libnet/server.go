package libnet

import (
	"io"
	"net"
	"strings"
	"time"

	"github.com/wuqifei/server_lib/concurrent"
	"github.com/wuqifei/server_lib/libnet/def"
	"github.com/wuqifei/server_lib/libnet/libsession"
	"github.com/wuqifei/server_lib/libnet/message"
	"github.com/wuqifei/server_lib/logs"
)

type Server struct {
	clientGroup *concurrent.ConcurrentIDGroupMap
	listerer    net.Listener
	protocol    def.Protocol
	Options     *ServerOptions
	OnClose     func(sessID uint64)
	OnConnect   func(sessID uint64)
}

func NewServer(l net.Listener, p def.Protocol) *Server {
	return &Server{
		clientGroup: concurrent.NewCocurrentIDGroup(),
		listerer:    l,
		protocol:    p,
	}
}

func (server *Server) Listener() net.Listener {
	return server.listerer
}

// 启动并且监听端口
func (server *Server) Run() {
	logs.Informational("libnet:Start to listen")

	for {
		sess, err := server.accept() //接收客户端连接
		if err != nil {
			//record
			logs.Error("libnet:Client Connect Failed sessionid(%v),error(%v)", sess.Get(libsession.SessionIDKey), err)
			continue
		}
		logs.Info("libnet:receive a client connect sessionid(%v)", sess.Get(libsession.SessionIDKey))
		sess.OnClose(server.sessionClosedCallback)
		sess.OnRecv(server.serssionRecvDataCallback)
	}
}

func (server *Server) accept() (libsession.Session, error) {
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
		iConn := server.protocol.NewConn(conn)
		sess := libsession.New(iConn, server.Options.SessionOption)
		sessID := (sess.Get(libsession.SessionIDKey)).(uint64)
		server.clientGroup.Set(sessID, sess)
		if server.OnConnect != nil {
			server.OnConnect(sessID)
		}
		return sess, nil
	}
}

// tcp服务关闭
func (s *Server) Stop() {
	s.listerer.Close()
	//释放所有的连接
	s.clientGroup.Dispose()
}

func (s *Server) sessionClosedCallback(sess libsession.Session) {
	sessID := (sess.Get(libsession.SessionIDKey)).(uint64)

	se := s.clientGroup.Get(sessID)
	if se != nil && se.(libsession.Session) == sess {

		s.clientGroup.Del(sessID)
		if s.OnClose != nil {
			s.OnClose(sessID)
		}
	}

	if se != nil {
		logs.Debug("libnet:onclose callback: sess(%l)", sessID)
	}
}

func (s *Server) serssionRecvDataCallback(msg *def.LibnetMessage, sess libsession.Session) {
	defer func() {
		if r := recover(); r != nil {
			//这里并不信任逻辑层的注册以及客户端的输入
			logs.Error("libnet: has recevice a panic(%v)", r)
		}
	}()

	sessID := (sess.Get(libsession.SessionIDKey)).(uint64)
	handler := message.GetHandler(msg.MsgID)

	//执行回调方法
	ackData, ackErr := handler(msg.Content, sessID)
	if ackErr != nil {
		//服务器处理错误
		logs.Error("libnet:server ackdata handler error (%v)", ackErr)
		return
	}

	// 空值直接返回
	if ackData == nil {
		return
	}

	length := len(ackData)
	if length != 2 {
		//服务器错误
		logs.Error("libnet:server ackdata length error msgID(%d),sessionId(%d)", msg.MsgID, sessID)
		return
	}

	resMsg := &def.LibnetMessage{}

	resMsg.MsgID = ackData[0].(uint16)
	resMsg.Content = ackData[1].([]byte)

	sess.Send(resMsg)
}

// 注册路由
func (s *Server) RegistRoute(msgType uint16, ret def.MessageHandlerWithRet) {
	logs.Info("libnet:registe route (%d)", msgType)
	message.Register(msgType, ret)
}

// 禁用某个用户
func (s *Server) DisableSession(sessID uint64) error {
	sessInterface := s.clientGroup.Get(sessID)
	if sessInterface == nil {
		logs.Error("libnet:DisableSession error(%v)", def.SessionCannotFoundErr)
		return def.SessionCannotFoundErr
	}
	sess := sessInterface.(libsession.Session)
	err := sess.Close()
	return err
}

// 发送确定的信息，给指定的一群session用户
func (s *Server) BroadCastConstMsg(groups []uint64, msg *def.LibnetMessage, failed chan<- []uint64) {

	failedUser := []uint64{}

	for _, id := range groups {

		if err := s.SendMessage2Sess(id, msg); err != nil {
			failedUser = append(failedUser, id)
		}
	}

	//将失败的用户返回去
	failed <- failedUser
}

func (s *Server) Check(sessID uint64) bool {
	sessInterface := s.clientGroup.Get(sessID)
	if sessInterface == nil {
		return false
	}
	return true
}

func (s *Server) SendMessage2Sess(sessID uint64, msg *def.LibnetMessage) error {

	sessInterface := s.clientGroup.Get(sessID)
	if sessInterface == nil {
		logs.Error("libnet:DisableSession error(%v)", def.SessionCannotFoundErr)
		return def.SessionCannotFoundErr
	}
	sess := sessInterface.(libsession.Session)
	return sess.Send(msg)
}
