package libnet

import (
	"io"
	"net"
	"strings"
	"time"

	"github.com/wuqifei/server_lib/concurrent"
	"github.com/wuqifei/server_lib/libnet/def"
	"github.com/wuqifei/server_lib/libnet/message"
	"github.com/wuqifei/server_lib/libnet/session"
	"github.com/wuqifei/server_lib/logs"
)

type Server struct {
	clientGroup  *concurrent.ConcurrentIDGroupMap
	listerer     net.Listener
	protocol     def.Protocol
	Options      *ServerOptions
	onClose 	func(sessID uint64)()
}

func NewServer(l net.Listener, p def.Protocol) *Server {
	return &Server{
		clientGroup:  concurrent.NewCocurrentGroup(),
		listerer:     l,
		protocol:     p,
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
			logs.Error("libnet:Client Connect Failed sessionid(%d),error(%v)", sess.ID(), err)
			continue
		}
		logs.Informational("libnet:receive a client connect sessionid(%d)", sess.ID())
		sess.AddCloseCallback(server.sessionClosedCallback)
		sess.AddRecvCallBack(server.serssionRecvDataCallback)
	}
}

func (server *Server) accept() (*session.Session, error) {
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
		codec := server.protocol.NewCodec(conn)
		session := session.NewSession(codec, server.Options.ReadTimeOutTimes, server.Options.SendQueueBuf, server.Options.RecvQueueBuf, server.Options.RecvTimeOut, server.Options.SendTimeOut)
		server.clientGroup.Set(session.ID(), session)
		return session, nil
	}
}

// tcp服务关闭
func (s *Server) Stop() {
	s.listerer.Close()
	//释放所有的连接
	s.clientGroup.Dispose()
}

func (s *Server) sessionClosedCallback(sess *session.Session) {	
	// 删除这个session
	se := s.clientGroup.Get(sess.ID())
	if s != nil && se.(*session.Session) == sess {

		s.clientGroup.Del(sess.ID())
		s.onClose(sess.ID())
	}

	if se != nil {
		logs.Debug("libnet:onclose callback: (%l), se(%l)",sess.ID(),se.(*session.Session).ID())
	}
}

func (s *Server) serssionRecvDataCallback(data interface{}, msgID uint16, sess *session.Session, err error) {
	defer func() {
		if r := recover(); r != nil {
			//这里并不信任逻辑层的输入
			logs.Error("libnet: has recevice a panic(%v)", r)
		}
	}()

	if err != nil {
		//记录处理，不返回
		logs.Error("libnet:server returned a error msgID(%d),error(%v),sessionId(%d)", msgID, err, sess.ID())
		return
	}
	handler := message.GetHandler(msgID)

	//执行回调方法
	ackData := handler(data.([]byte), sess.ID())

	length := len(ackData)
	if length != 2 {
		//服务器错误
		logs.Error("libnet:server ackdata length error msgID(%d),sessionId(%d)", msgID, sess.ID())
		return
	}

	var packet []byte
	resID := ackData[0].(uint16)
	resData := ackData[1].([]byte)
	packet = sess.PackData(resID, resData)

	if len(packet) > 0 {
		//如果有数据 则发送
		sess.Send(packet)
	}
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
	sess := sessInterface.(*session.Session)
	err := sess.Close()
	return err
}

// 发送确定的信息，给指定的session用户
func (s *Server) BroadCastConstMsg(groups []uint64, msg def.LibnetMessage, failed chan <- []uint64) {

	failedUser := []uint64{}

	for _, id := range groups {

		if err := s.SendMessage2Sess(id, msg);err != nil {

			failedUser = append(failedUser, id)
		}
	}

	//将失败的用户返回去
	failed <-  failedUser
}

func (s *Server) SendMessage2Sess(sessID uint64, msg def.LibnetMessage) error{

	sessInterface := s.clientGroup.Get(sessID)
	if sessInterface == nil {
		logs.Error("libnet:DisableSession error(%v)", def.SessionCannotFoundErr)
		return def.SessionCannotFoundErr
	}
	sess := sessInterface.(*session.Session)
	data := sess.PackData(msg.MsgID, msg.Content)
	return sess.Send(data)
}

// 设置session关闭的回调
func (s *Server)OnClose(callback func(sessID uint64)()) {
	if callback == nil {

		logs.Error("libnet:OnClose set nil")
		panic("libnet:OnClose set nil")
	}
	s.onClose = callback
}