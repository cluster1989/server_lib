package libsession

import (
	"context"
	"net"
	"sync"
)

// session的配置
type Options struct {
	MaxAge int //最大的存在秒数
}

// Session的接口
type Session interface {
	Get(key interface{}) interface{}
	Set(key interface{}, val interface{})
	Delete(key interface{}) //并非真正删除
	Clear()                 //这里是真正的清除
	Options(Options)
	Save() error
}

type session struct {
	ID     uint64
	CTX    context.Context
	Cancel context.CancelFunc
	conn   net.Conn
	sync.Mutex
}

func New() Session {
	sess := &session{}
	sess.CTX, sess.Cancel = context.WithCancel(context.Background())
	return sess
}

func (s *session) Get(key interface{}) interface{} {
	s.Lock()
	defer s.Unlock()
	return s.CTX.Value(key)
}

func (s *session) Set(key interface{}, val interface{}) {
	s.Lock()
	defer s.Unlock()
	context.WithValue(s.CTX, key, val)
}

func (s *session) Delete(key interface{}) {
	context.WithValue(s.CTX, key, nil)
}

func (s *session) Clear() {

}

func (s *session) Options(Options) {
}

func (s *session) Save() error {
}
