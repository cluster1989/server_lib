package session

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/wqf/common_lib/concurrent"
	"github.com/wqf/common_lib/libio"
	"github.com/wqf/common_lib/libnet/def"
)

var SessionClosedError = errors.New("Session Closed")

var globalSessionId *concurrent.AtomicUint64

type Session struct {
	onRecv        func(data interface{}, msgId uint16, sess *Session, err error)
	id            uint64
	codec         def.Codec
	sendChan      chan interface{}
	recvChan      chan interface{}
	closeFlag     *concurrent.AtomicBoolean
	closeChan     chan int
	closeMutex    sync.Mutex
	closeCallback func(s *Session)
	once          *sync.Once

	readTimeOut  time.Duration //读取超时
	writeTimeOut time.Duration //写入超时
}

func init() {
	globalSessionId = concurrent.NewAtomicUint64(0)
}

func NewSession(codec def.Codec, sendChanSize, recvChanSize int, readTimeOut, writeTimeOut time.Duration) *Session {
	return newSession(codec, sendChanSize, recvChanSize, readTimeOut, writeTimeOut)
}

func newSession(codec def.Codec, sendChanSize int, recvChanSize int, readTimeOut, writeTimeOut time.Duration) *Session {
	session := &Session{}
	session.codec = codec
	session.id = globalSessionId.IncrementAndGet()
	session.closeFlag = concurrent.NewAtomicBoolean(false) //默认未关闭

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

	session.writeTimeOut = writeTimeOut
	session.readTimeOut = readTimeOut
	session.sendChan = make(chan interface{}, sendChanSize)
	session.recvChan = make(chan interface{}, recvChanSize)

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
		select {
		case msg := <-s.recvChan:
			if s.onRecv != nil {
				data, msgID, err := s.parse(msg)
				s.onRecv(data, msgID, s, err)
			}
		case <-time.After(s.readTimeOut):
			//超时
			return
		}
		//设置超市时间
	}
}

func (s *Session) sendChanLoop() {
	defer s.Close()
	for {
		select {
		case msg := <-s.sendChan:
			if s.codec.Send(msg) != nil {
				return
			}
		case <-s.closeChan:
			return
		case <-time.After(s.writeTimeOut):
			//超时
			return
		}
	}
}

func (s *Session) recvLoop() {
	defer s.Close()

	for {
		data, err := s.codec.Receive()
		if err != nil {
			//记录错误
			continue
		}
		if data == nil {
			//数据错误，直接关闭
			s.Close()
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
	}
	//这里设置发送超时
	select {
	case s.sendChan <- msg:
		return nil

	}
}

//一个session关闭只能实现一次
func (s *Session) close() error {
	if s.IsClosed() {
		return nil
	}
	var err error
	s.once.Do(func() {

		err = s.codec.Close() //关闭连接
		close(s.closeChan)    //关闭通道
		close(s.recvChan)
		close(s.sendChan)
		s.invokeCloseCallBack() //发送关闭的回调
		s.closeFlag.Set(true)   //设置已关闭
	})

	return err
}

//关闭session
func (s *Session) Close() error {
	s.closeChan <- 1
	return nil
}

func (s *Session) parse(data interface{}) (mdata interface{}, msgId uint16, err error) {
	//解析对象
	reqData := mdata.([]byte)
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
