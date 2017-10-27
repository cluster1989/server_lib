package libsession

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/wuqifei/server_lib/logs"

	"github.com/wuqifei/server_lib/concurrent"
	"github.com/wuqifei/server_lib/libio"

	"github.com/wuqifei/server_lib/libnet/def"
)

// 接收到数据
type OnRecv func(msg *def.LibnetMessage, sess Session)

// 关闭session
type OnClose func(sess Session)

var globalSessionId *concurrent.AtomicUint64

// session的配置
type Options struct {
	ReadTimeout time.Duration //读取的超时

	// 允许的超时次数
	ReadTimeoutTimes int

	RecvChanSize int //接收和发送队列的大小
	SendChanSize int

	IsLittleEndian bool
}

// Session的接口
type Session interface {
	Get(key interface{}) interface{}
	Set(key interface{}, val interface{})
	Delete(key interface{}) //删除存储的数据
	Clear()                 //直接清除

	OnRecv(recv OnRecv)
	OnClose(close OnClose)
	Send(msg *def.LibnetMessage) error
	Close() error
}

type session struct {
	id     uint64
	Params *concurrent.ConcurrentMap
	conn   def.Conn
	sync.Mutex

	//已经超时的次数
	timeoutTimes *concurrent.AtomicInt32
	options      Options
	token        string

	//释放的时候，为释放服务
	disposeOnce sync.Once

	onRecv  OnRecv
	onClose OnClose

	closeFlag *concurrent.AtomicBoolean

	recvChan chan interface{}
	sendChan chan interface{}
}

func init() {
	globalSessionId = concurrent.NewAtomicUint64(0)
}

func New(conn def.Conn, option Options) Session {
	sess := &session{}
	sess.conn = conn
	sess.Params = concurrent.NewCocurrentMap()
	sess.id = globalSessionId.IncrementAndGet()
	//新建之后，直接把id set进去
	sess.Params.Set(SessionIDKey, sess.id)
	sess.options = option
	sess.closeFlag = concurrent.NewAtomicBoolean(false)
	sess.check()
	sess.Run()
	return sess
}

func (s *session) check() {
	if s.options.ReadTimeout == 0 {
		s.options.ReadTimeout = time.Duration(60) * time.Second
	}
}

func (s *session) Run() {
	s.recvChan = make(chan interface{}, s.options.RecvChanSize)
	s.timeoutTimes = concurrent.NewAtomicInt32(0)
	s.sendChan = make(chan interface{}, s.options.SendChanSize)

	go s.chanLoop()
	go s.recvLoop()

}

func (s *session) chanLoop() {
	defer s.Close()
	for {

		select {
		case msg := <-s.recvChan:

			if s.isClose() {
				//关闭 让出当前去程
				logs.Debug("session: recv chan loop hasclosed id(%d)", s.id)
				runtime.Goexit()
				return
			}

			model, e := s.parse(msg)
			if e != nil {
				return
			}
			s.invokeRecv(model)
			s.timeoutTimes.Set(0)

		case msg := <-s.sendChan:

			if s.isClose() {
				logs.Debug("session: send chan loop hasclosed id(%d)", s.id)
				runtime.Goexit()
				return
			}
			logs.Debug("session: send session message message(%v) session(%d)", msg, s.id)
			if err := s.conn.Send(msg); err != nil {
				logs.Error("session: send session message error(%v) session(%d)", err, s.id)
				return
			}

		case <-time.After(s.options.ReadTimeout):
			t := s.timeoutTimes.IncrementAndGet()
			if t > int32(s.options.ReadTimeoutTimes) {
				logs.Info("session:has not send msg for (%d) times session(%d)", t, s.id)
				return
			}
			logs.Emergency("session:recv session message  timeout session(%d)", s.id)
		}
	}
}

func (s *session) recvLoop() {
	defer s.Close()
	for {
		data, err := s.conn.Receive()

		logs.Debug("session: recv session message len[%d] message(%v) session(%d)", len(data), data, s.id)
		if err != nil {
			//记录错误

			if err == io.EOF || err == io.ErrUnexpectedEOF {
				logs.Info("session: session has closed [%d])", err, s.id)
				return
			}

			logs.Error("session: recv session message error(%v) session(%d))", err, s.id)
			continue
		}

		//写在这里的原因是不希望代码再往下走
		if s.isClose() {
			logs.Debug("session: recvloop hasclosed id(%d)", s.id)
			runtime.Goexit()
			return
		}

		s.recvChan <- data
	}
}

func (s *session) Get(key interface{}) interface{} {
	return s.Params.Get(key)
}

func (s *session) Set(key interface{}, val interface{}) {
	s.Params.Set(key, val)
}

func (s *session) Delete(key interface{}) {
	s.Params.Del(key)
}

func (s *session) Clear() {
	//直接全部释放
	s.Params.Dispose()
	s.Params = concurrent.NewCocurrentMap()
}

func (s *session) OnRecv(recv OnRecv) {
	s.onRecv = recv
}

func (s *session) OnClose(close OnClose) {
	s.onClose = close
}

func (s *session) isClose() bool {
	return s.closeFlag.Get()
}

func (s *session) invokeClose() {
	if s.onClose == nil || s.isClose() {
		return
	}
	s.Lock()
	s.onClose(s)
	s.Unlock()
}

func (s *session) invokeRecv(msg *def.LibnetMessage) {
	if s.onClose == nil || s.isClose() {
		return
	}
	s.Lock()
	s.onRecv(msg, s)
	s.Unlock()
}

func (s *session) Send(msg *def.LibnetMessage) error {
	if s.isClose() {
		return fmt.Errorf("session: session is closed")
	}
	if s.sendChan == nil {
		return s.conn.Send(msg)
	}
	b := s.packData(msg)
	s.sendChan <- b
	return nil
}

func (s *session) packData(msg *def.LibnetMessage) interface{} {

	packet := make([]byte, 2)
	if s.options.IsLittleEndian {
		libio.PutUint16LE(packet, msg.MsgID)
	} else {
		libio.PutUint16BE(packet, msg.MsgID)
	}
	packet = append(packet, msg.Content...)
	return packet
}

func (s *session) close() error {
	if s.isClose() {
		return nil
	}
	var err error
	s.disposeOnce.Do(func() {
		s.closeFlag.Set(true)
		s.conn.Close() //关闭连接
		close(s.recvChan)
		close(s.sendChan)
		logs.Info("session: already closed session(%d)", s.id)
		s.invokeClose()
	})
	return err
}

func (s *session) Close() error {
	if s.isClose() {
		return nil
	}
	logs.Info("session:session send Close message session(%d)", s.id)
	return s.close()
}

func (s *session) parse(data interface{}) (msg *def.LibnetMessage, err error) {
	bData, ok := data.([]byte)
	if !ok || len(bData) < 2 {
		return nil, fmt.Errorf("session:parse data is not a []byte ,[%v],[%v]", data, reflect.TypeOf(data))
	}
	cmdByte := bData[0:2]
	if cmdByte == nil {
		return nil, fmt.Errorf("session:parse data header is not a length ,[%v],[%v]", data, reflect.TypeOf(data))
	}
	msg = &def.LibnetMessage{}
	if s.options.IsLittleEndian {
		msg.MsgID = libio.GetUint16LE(cmdByte)
	} else {
		msg.MsgID = libio.GetUint16BE(cmdByte)
	}
	msg.Content = bData[2:]
	return
}
