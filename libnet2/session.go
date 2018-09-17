package libnet2

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/wuqifei/server_lib/concurrent"
	"github.com/wuqifei/server_lib/libio"
)

var globalSessionId *concurrent.AtomicUint64

type defaultSession struct {
	// session配置
	option *SessionOption2
	id     uint64
	params *concurrent.ConcurrentMap
	conn   net.Conn
	//已经超时的次数
	timeoutTimes *concurrent.AtomicInt32

	//释放的时候，为释放服务,保证一个session只被释放一次
	disposeOnce sync.Once

	onClose OnSessClose
	onError OnSessError
	onRecv  OnSessRecv

	closeFlag *concurrent.AtomicBoolean

	recvChan  chan []byte
	sendChan  chan []byte
	closeChan chan bool

	reader *libio.Reader
	writer *libio.Writer
}

func init() {
	globalSessionId = concurrent.NewAtomicUint64(0)
}

func newDefaultSession(conn net.Conn, sessionOption *SessionOption2) *defaultSession {
	sess := new(defaultSession)
	sess.option = sessionOption
	sess.check()
	sess.conn = conn
	sess.params = concurrent.NewCocurrentMap()
	sess.id = globalSessionId.IncrementAndGet()
	sess.timeoutTimes = concurrent.NewAtomicInt32(0)
	sess.closeFlag = concurrent.NewAtomicBoolean(false)
	sess.recvChan = make(chan []byte, sess.option.RecvChanSize)
	sess.sendChan = make(chan []byte, sess.option.SendChanSize)
	sess.closeChan = make(chan bool, 1)
	sess.reader = libio.NewReader(conn)
	sess.writer = libio.NewWriter(conn)
	return sess
}

func (s *defaultSession) check() {
	if s.option.ReadTimeout == 0 {
		s.option.ReadTimeout = time.Duration(60) * time.Second
	}
	if s.option.RecvChanSize < 1 {
		s.option.RecvChanSize = 1
	}
	if s.option.SendChanSize < 1 {
		s.option.SendChanSize = 1
	}
}

//  发送数据
func (s *defaultSession) Send(val []byte) error {
	if ServerPacket == nil {
		panic(errors.New("服务的解析对象不能为nil"))
	}
	if s.option.SendChanSize > 1 {
		s.sendChan <- val
	} else {
		return ServerPacket.Write(s.writer, val)
	}
	return nil
}

// 关闭
func (s *defaultSession) Close() error {
	// 已经关闭
	if s.closeFlag.Get() {
		return nil
	}
	s.closeChan <- true
	return nil
}

// 收到信息
func (s *defaultSession) Recv(onRecv OnSessRecv) {
	s.onRecv = onRecv
}

// 设置参数
func (s *defaultSession) Set(key, val interface{}) error {
	if key == nil || val == nil {
		return ErrValueNull
	}
	s.params.Set(key, val)
	return nil
}

// 获取参数
func (s *defaultSession) Get(key interface{}) (interface{}, error) {
	if key == nil {
		return nil, ErrValueNull
	}
	val := s.params.Get(key)
	return val, nil
}

// 删除参数
func (s *defaultSession) Del(key interface{}) (bool, error) {
	if key == nil {
		return false, ErrValueNull
	}

	s.params.Del(key)
	return true, nil
}

// 清空参数
func (s *defaultSession) Clear() error {
	//直接全部释放
	s.params.Dispose()
	s.params = concurrent.NewCocurrentMap()
	return nil
}

// 获取连接
func (s *defaultSession) GetConn() net.Conn {
	return s.conn
}

func (s *defaultSession) Reader() *libio.Reader {
	return s.reader
}
func (s *defaultSession) Writer() *libio.Writer {
	return s.writer
}

// 获取内存存储的uniquid
func (s *defaultSession) GetUniqueID() uint64 {
	return s.id
}

// 刷新session的id
func (s *defaultSession) setUniqueID(id uint64) {
	s.id = id
}

func (s *defaultSession) Accept() {
	go s.chanLoop()
	go s.recvLoop()
}

func (s *defaultSession) chanLoop() {
	defer s.Close()
	for {
		select {
		case msg, flag := <-s.recvChan:
			if !flag {
				return
			}

			s.onRecv(s, msg)
			s.timeoutTimes.Set(0)

		case msg, flag := <-s.sendChan:

			if !flag {
				return
			}

			ServerPacket.Write(s.writer, msg)

		case <-time.After(s.option.ReadTimeout):
			t := s.timeoutTimes.IncrementAndGet()
			if t > int32(s.option.ReadTimeoutTimes) {
				//直接关闭
				return
			}

		case flag := <-s.closeChan:
			if flag {
				// 如果是关闭
				s.close()
			}
		}
	}
}

func (s *defaultSession) recvLoop() {

	defer s.Close()
	for {
		if ServerPacket == nil {
			panic(errors.New("服务的解析对象不能为nil"))
		}

		data, err := ServerPacket.Read(s.reader)
		if err != nil {

			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// 这里表示session已经关闭

				return
			}

			if s.onError != nil {
				s.onError(s, err)
			}
			continue
		}

		if data == nil {
			continue
		}

		if s.option.RecvChanSize > 1 {
			s.recvChan <- data
		} else {
			s.onRecv(s, data)
		}
	}
}

func (s *defaultSession) close() error {
	var err error

	s.closeFlag.Set(true)
	s.disposeOnce.Do(func() {
		s.conn.Close() //关闭连接
		close(s.closeChan)
		close(s.recvChan)
		close(s.sendChan)
	})
	s.onClose(s)
	return err
}
