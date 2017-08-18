package libnet

import (
	"io"
	"net"
	"strings"
	"time"

	"github.com/wqf/common_lib/libio"
	"github.com/wqf/common_lib/libtime"

	"github.com/wqf/common_lib/concurrent"

	"github.com/wqf/common_lib/libnet/def"
	"github.com/wqf/common_lib/libnet/message"
	"github.com/wqf/common_lib/libnet/session"
)

type Server struct {
	sessionGroup *concurrent.SyncGroupMap
	listerer     net.Listener
	protocol     def.Protocol
	Options      *ServerOptions
	timeWheel    *libtime.TimerWheel
}

func NewServer(l net.Listener, p def.Protocol) *Server {
	return &Server{
		sessionGroup: concurrent.NewGroup(),
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
		sess, err := server.Accept() //接收客户端连接
		if err != nil {
			//record

			continue
		}
		server.sessionGroup.Set(sess.ID(), sess)
		sess.AddCloseCallback(server.sessionClosedCallback)
		sess.AddRecvCallBack(server.serssionRecvDataCallback)
	}
}

func (server *Server) Accept() (*session.Session, error) {
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
		session := session.NewSession(codec, server.Options.SendQueueBuf, server.Options.RecvQueueBuf, server.Options.RecvTimeOut, server.Options.SendTimeOut)
		return session, nil
	}
}

func (s *Server) Stop() {
	s.listerer.Close()
	s.sessionGroup.Dispose()
}

func (s *Server) sessionClosedCallback(sess *session.Session) {
	//删除id
	s.sessionGroup.Del(sess.ID())
}

func (s *Server) serssionRecvDataCallback(data interface{}, msgId uint16, sess *session.Session, err error) {
	handler := message.GetHandler(msgId)
	ackCodec, err := handler(data.([]byte))
	if err != nil {
		//记录处理，不返回
		return
	}
	ackData := ackCodec.MessageSerialize()
	ackID := ackCodec.MessageType()
	var packet []byte
	if len(ackData) == 1 {
		packet = s.packData(ackData[0].([]byte), msgId)
	} else if len(ackData) == 2 {
		packet = s.packData(ackData[0].([]byte), msgId)
		//组合session
	}
	if len(packet) > 0 {
		//如果有数据 则发送
		sess.Send(packet)
	}
}

func (s *Server) packData(data []byte, msgId uint16) []byte {

	packet := make([]byte, 2)
	if s.Options.IsLittleIndian {
		libio.PutUint16LE(packet, msgId)
	} else {
		libio.PutUint16BE(packet, msgId)
	}
	packet = append(packet, data...)

	return packet
}
