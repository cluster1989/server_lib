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
	"github.com/wuqifei/server_lib/libtime"
)

type Server struct {
	clientGroup  *concurrent.SyncGroupMap
	reflectGroup *concurrent.SyncUint64GroupMap
	listerer     net.Listener
	protocol     def.Protocol
	Options      *ServerOptions
	timeWheel    *libtime.TimerWheel
}

func NewServer(l net.Listener, p def.Protocol) *Server {
	return &Server{
		clientGroup:  concurrent.NewGroup(),
		reflectGroup: concurrent.NewUint64Group(),
		timeWheel:    libtime.NewTimerWheel(),
		listerer:     l,
		protocol:     p,
	}
}

func (server *Server) Listener() net.Listener {
	return server.listerer
}

func (server *Server) Run() {
	for {
		sess, err := server.accept() //接收客户端连接
		if err != nil {
			//record

			continue
		}
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
		server.DoHeartTask(session)
		return session, nil
	}
}

func (s *Server) DoHeartTask(sess *session.Session) {

	task := sess.SetupHeartTask()
	taskID := s.timeWheel.AddTask(s.Options.HeartBeatTime, -1, task)

	sess.HeartTaskID = taskID
}

func (s *Server) Stop() {
	s.listerer.Close()
	s.timeWheel.Stop()
	s.clientGroup.Dispose()
	s.clientGroup.Dispose()
}

func (s *Server) sessionClosedCallback(sess *session.Session) {
	//删除id
	userID := s.reflectGroup.Get(sess.ID())
	s.timeWheel.CancelTimer(sess.HeartTaskID)

	if userID == nil {
		return
	}
	s.clientGroup.Del(userID.(uint64))
	s.reflectGroup.Del(sess.ID())
}

func (s *Server) serssionRecvDataCallback(data interface{}, msgID uint16, sess *session.Session, err error) {
	if err != nil {
		//记录处理，不返回
		return
	}
	handler := message.GetHandler(msgID)

	isWildMsg := !s.querySessionIDISExists(sess.ID())

	ackData := handler(data.([]byte), isWildMsg)

	length := len(ackData)
	if length < 2 {
		//服务器错误
		return
	}

	var packet []byte
	if length >= 2 {
		resID := ackData[0].(uint16)
		resData := ackData[1].([]byte)

		packet = sess.PackData(resID, resData)

		if length == 3 {
			//组合session
			uniqueID := ackData[2].(uint64)
			//与session一起组合
			s.clientGroup.Set(uniqueID, sess)
			s.reflectGroup.Set(sess.ID(), uniqueID)
		}
	}
	if len(packet) > 0 {
		//如果有数据 则发送
		sess.Send(packet)
	}
}

func (s *Server) RegistRoute(msgType uint16, ret def.MessageHandlerWithRet) {
	message.Register(msgType, ret)
}

func (s *Server) RegistHeartBeat(msgType uint16, ret def.MessageHandlerWithRet) {
	message.RegisterHeartBeat(msgType, ret)
}

func (s *Server) QueryUserIDIsExists(uniqueID uint64) bool {
	sess := s.clientGroup.Get(uniqueID)
	if sess == nil {
		return false
	} else {
		return true
	}
}

func (s *Server) querySessionIDISExists(sessID uint64) bool {
	uniqueID := s.reflectGroup.Get(sessID)
	if uniqueID != nil {
		return true
	}
	return false

}

func (s *Server) CloseUser(userID uint64) {
	if !s.QueryUserIDIsExists(userID) {
		return
	}

	sess := s.clientGroup.Get(userID).(*session.Session)
	sess.Close()
}

// 一样第一个是msgid ,第二个msg data
func (s *Server) BroadCastMsg(groups []uint64, ack []interface{}) {
	for _, id := range groups {
		if !s.QueryUserIDIsExists(id) {
			continue
		}
		sess := s.clientGroup.Get(id).(*session.Session)
		data := sess.PackData(ack[0].(uint16), ack[1].([]byte))
		err := sess.Send(data)
		if err != nil {
			//记录
		}
	}

}
