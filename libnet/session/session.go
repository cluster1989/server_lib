package session

import (
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/wuqifei/server_lib/concurrent"
	"github.com/wuqifei/server_lib/libio"
	"github.com/wuqifei/server_lib/libnet/def"
	"github.com/wuqifei/server_lib/logs"
)

var SessionClosedError = errors.New("Session Closed")
var SessionMsgNoSuchTypeError = errors.New("No such type ")

var globalSessionId *concurrent.AtomicUint64

type Session struct {
	onRecv        func(data interface{}, msgId uint16, sess *Session, err error)
	id            uint64
	conn          def.Conn
	sendChan      chan interface{}
	recvChan      chan interface{}
	closeFlag     *concurrent.AtomicBoolean
	closeMutex    sync.Mutex
	closeCallback func(s *Session)
	once          *sync.Once

	readTimeOut      time.Duration           //读取超时
	writeTimeOut     time.Duration           //写入超时
	readTimeOutTimes int                     // 允许超时次数
	timeOutTimes     *concurrent.AtomicInt32 //已经超时次数
}

func init() {
	globalSessionId = concurrent.NewAtomicUint64(0)
}

func NewSession(conn def.Conn, readTimeOutTimes, sendChanSize, recvChanSize int, readTimeOut, writeTimeOut time.Duration) *Session {
	return newSession(conn, readTimeOutTimes, sendChanSize, recvChanSize, readTimeOut, writeTimeOut)
}

func newSession(conn def.Conn, readTimeOutTimes, sendChanSize int, recvChanSize int, readTimeOut, writeTimeOut time.Duration) *Session {

	session := &Session{}
	session.conn = conn
	session.id = globalSessionId.IncrementAndGet()
	session.closeFlag = concurrent.NewAtomicBoolean(false) //默认未关闭
	session.once = &sync.Once{}

	//错误检查
	if sendChanSize == 0 {
		sendChanSize = 10
	}
	if recvChanSize == 0 {
		recvChanSize = 10
	}
	if readTimeOut == 0 {
		readTimeOut = time.Duration(60) * time.Second
	}
	if writeTimeOut == 0 {
		writeTimeOut = time.Duration(60) * time.Second
	}

	if readTimeOutTimes == 0 {
		readTimeOutTimes = 3
	}

	session.writeTimeOut = writeTimeOut
	session.readTimeOut = readTimeOut
	session.sendChan = make(chan interface{}, sendChanSize)
	session.recvChan = make(chan interface{}, recvChanSize)
	session.readTimeOutTimes = readTimeOutTimes

	session.timeOutTimes = concurrent.NewAtomicInt32(0)

	go session.recvChanLoop()
	go session.sendChanLoop()
	go session.recvLoop()

	return session
}

func (s *Session) ID() uint64 {
	return s.id
}

func (s *Session) Conn() def.Conn {
	return s.conn
}

func (s *Session) IsClosed() bool {
	return s.closeFlag.Get()
}

func (s *Session) recvChanLoop() {
	defer s.Close()
	for {
		if s.IsClosed() {

			//这里强制当前去程进入闲置
			logs.Debug("libnet:session recv chan loop hasclosed id(%d)", s.ID())
			runtime.Goexit()
			return
		}

		select {
		case msg := <-s.recvChan:
			if s.onRecv != nil {
				data, msgID, err := s.parse(msg)
				if err != nil {
					logs.Error("libnet:session parse message msgId(%d) data(%v) err(%v)", data, msgID, err)
				}
				s.onRecv(data, msgID, s, err)
			}
			s.timeOutTimes.Set(0)

		case <-time.After(s.readTimeOut):
			t := s.timeOutTimes.IncrementAndGet()

			if t > int32(s.readTimeOutTimes) {
				logs.Info("libnet:session has not send msg for (%d) times session(%d)", t, s.ID())
				return
			} else {
				logs.Emergency("libnet:session recv session message  timeout session(%d)", s.ID())
				continue
			}
		}
	}
}

func (s *Session) sendChanLoop() {
	defer s.Close()
	for {
		if s.IsClosed() {

			//这里强制当前去程进入闲置
			logs.Debug("libnet:session send chan loop hasclosed id(%d)", s.ID())
			runtime.Goexit()
			return
		}

		select {
		case msg := <-s.sendChan:
			logs.Debug("libnet:session send session message message(%v) session(%d)", msg, s.ID())
			if err := s.conn.Send(msg); err != nil {
				logs.Error("libnet:session send session message error(%v) session(%d)", err, s.ID())
				return
			}
		case <-time.After(s.writeTimeOut):
			//超时
			logs.Emergency("libnet:session send session message timeout session(%d)", s.ID())
			return
		}
	}
}

func (s *Session) recvLoop() {
	defer s.Close()

	for {
		if s.IsClosed() {

			//这里强制当前去程进入闲置
			logs.Debug("libnet:session recv loop hasclosed id(%d)", s.ID())
			runtime.Goexit()
			return
		}

		data, err := s.conn.Receive()
		logs.Debug("libnet:session recv session message message(%#v) session(%d)", data, s.ID())

		if err != nil {
			//记录错误
			logs.Error("libnet:session recv session message error(%v) session(%d))", err, s.ID())
			continue
		}

		if data == nil || len(data) == 0 {
			//数据错误，直接关闭
			logs.Emergency("libnet:session recv empty message and closed session(%d)", s.ID())
			return
		}
		if !s.IsClosed() {
			//将数据放入缓冲池中
			s.recvChan <- data
		}
	}
}

func (s *Session) Send(msg interface{}) error {
	if s.IsClosed() {
		return SessionClosedError
	}
	if s.sendChan == nil {
		return s.conn.Send(msg)
	} else {
		s.sendChan <- msg
	}
	return nil
}

//一个session关闭只能实现一次
func (s *Session) close() error {
	if s.IsClosed() {
		return nil
	}
	var err error
	s.once.Do(func() {

		s.closeFlag.Set(true) //设置已关闭

		close(s.recvChan)       //关闭接收通道
		close(s.sendChan)       //关闭发送通道
		s.conn.Close()          //关闭当前连接
		s.invokeCloseCallBack() //发送关闭的回调

		logs.Info("libnet:session already closed session(%d)", s.ID())
	})

	return err
}

//关闭session
func (s *Session) Close() error {

	logs.Info("libnet:session send Close message session(%d)", s.ID())
	return s.close()
}

func (s *Session) parse(data interface{}) (mdata interface{}, msgId uint16, err error) {

	//解析对象
	reqData, ok := data.([]byte)
	if !ok {
		err = SessionMsgNoSuchTypeError
		return
	}
	//先取协议号
	cmdByte := reqData[0:2]
	if cmdByte == nil {
		//记录错误,直接关闭，没有标识
		s.Close()
		return
	}
	cmdUint16 := libio.GetUint16BE(cmdByte)
	cmdObj := reqData[2:]
	mdata = cmdObj
	msgId = cmdUint16
	err = nil
	return
}

func (s *Session) AddCloseCallback(callback func(s *Session)) *Session {
	s.closeMutex.Lock()
	defer s.closeMutex.Unlock()
	s.closeCallback = callback
	return s
}

func (s *Session) AddRecvCallBack(callback func(data interface{}, msgId uint16, sess *Session, err error)) *Session {
	s.closeMutex.Lock()
	defer s.closeMutex.Unlock()
	s.onRecv = callback
	return s
}

func (s *Session) invokeCloseCallBack() {
	if s.IsClosed() || s.closeCallback == nil {
		return
	}
	s.closeMutex.Lock()
	defer s.closeMutex.Unlock()
	s.closeCallback(s)
}

func (s *Session) PackData(msgID uint16, data []byte) []byte {
	packet := make([]byte, 2)
	libio.PutUint16BE(packet, msgID)
	packet = append(packet, data...)
	return packet
}
