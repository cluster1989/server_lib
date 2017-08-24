package session

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/wuqifei/server_lib/concurrent"
	"github.com/wuqifei/server_lib/libio"
	"github.com/wuqifei/server_lib/libnet/def"
	"github.com/wuqifei/server_lib/libnet/message"
	"github.com/wuqifei/server_lib/libtime"
)

var SessionClosedError = errors.New("Session Closed")
var SessionMsgNoSuchTypeError = errors.New("No such type ")

var globalSessionId *concurrent.AtomicUint64

type Session struct {
	onRecv        func(data interface{}, msgId uint16, sess *Session, err error)
	id            uint64
	codec         def.Codec
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
	//心跳id
	HeartTaskID int64
	HeartTask   *libtime.TimerTaskTimeOut
}

func init() {
	globalSessionId = concurrent.NewAtomicUint64(0)
}

func NewSession(codec def.Codec, readTimeOutTimes, sendChanSize, recvChanSize int, readTimeOut, writeTimeOut time.Duration) *Session {
	return newSession(codec, readTimeOutTimes, sendChanSize, recvChanSize, readTimeOut, writeTimeOut)
}

func newSession(codec def.Codec, readTimeOutTimes, sendChanSize int, recvChanSize int, readTimeOut, writeTimeOut time.Duration) *Session {

	session := &Session{}
	session.codec = codec
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

func (s *Session) Codec() def.Codec {
	return s.codec
}

func (s *Session) IsClosed() bool {
	return s.closeFlag.Get()
}

func (s *Session) recvChanLoop() {
	defer s.Close()
	for {
		if s.IsClosed() {
			return
		}

		select {
		case msg := <-s.recvChan:
			if s.onRecv != nil {
				data, msgID, err := s.parse(msg)
				s.onRecv(data, msgID, s, err)
			}
			s.timeOutTimes.Set(0)

		case <-time.After(s.readTimeOut):
			t := s.timeOutTimes.IncrementAndGet()

			if t > int32(s.readTimeOutTimes) {
				fmt.Println("session is really timeout -")
				return
			} else {
				fmt.Printf("recv data timeout :%d readtimeouttimes:%d \n", t, s.readTimeOutTimes)
				continue
			}
		}
		//设置超市时间
	}
}

func (s *Session) sendChanLoop() {
	defer s.Close()
	for {
		if s.IsClosed() {
			return
		}

		select {
		case msg := <-s.sendChan:
			if s.codec.Send(msg) != nil {
				return
			}
		case <-time.After(s.writeTimeOut):
			//超时
			return
		}
	}
}

func (s *Session) recvLoop() {
	defer s.Close()

	for {
		if s.IsClosed() {
			return
		}

		data, err := s.codec.Receive()
		fmt.Printf("recv msg :%v\n", data)
		if err != nil {
			//记录错误
			continue
		}
		if data == nil || len(data) == 0 {
			//数据错误，直接关闭
			return
		}
		s.recvChan <- data
		//将数据放入缓冲池中
	}
}

func (s *Session) Send(msg interface{}) error {
	if s.IsClosed() {
		return SessionClosedError
	}
	if s.sendChan == nil {
		return s.codec.Send(msg)
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

		close(s.recvChan)
		close(s.sendChan)
		s.codec.Close()         //关闭连接
		s.invokeCloseCallBack() //发送关闭的回调
		s.closeFlag.Set(true)   //设置已关闭

	})

	return err
}

//关闭session
func (s *Session) Close() error {

	return s.close()
}

func (s *Session) parse(data interface{}) (mdata interface{}, msgId uint16, err error) {

	//解析对象
	reqData, ok := mdata.([]byte)
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
	fmt.Println("%d,%v", cmdUint16, cmdObj)
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

func (s *Session) SetupHeartTask() (task *libtime.TimerTaskTimeOut) {

	if s.HeartTask == nil {

		s.closeMutex.Lock()
		task := &libtime.TimerTaskTimeOut{}
		task.Content = nil
		task.Callback = func(backData interface{}) {
			s.sendHeartMsg()
		}

		s.HeartTask = task
		s.closeMutex.Unlock()
	}
	return s.HeartTask
}

func (s *Session) sendHeartMsg() {
	handler := message.GetHeartBeatHandler()
	ackData := handler(nil, true)
	heartMsgID := ackData[0].(uint16)
	heartData := ackData[1].([]byte)
	packet := s.PackData(heartMsgID, heartData)
	err := s.Send(packet)
	if err != nil {
		//heart msg
	}

}
