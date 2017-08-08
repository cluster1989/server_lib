package libnet

import (
	"container/list"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var SessionClosedError = errors.New("Session Closed")
var SessionBlockedError = errors.New("Session Blocked")

var globalSessionId uint64

type Session struct {
	id             uint64
	codec          Codec
	manager        *Manager
	sendChan       chan interface{}
	closeFlag      int32
	closeChan      chan int
	closeMutex     sync.Mutex
	closeCallbacks *list.List
	Extra          interface{}
}

func NewSession(codec Codec, sendChanSize int) *Session {
	return newSession(nil, codec, sendChanSize)
}

func newSession(manager *Manager, codec Codec, sendChanSize int) *Session {
	session := &Session{
		codec:     codec,
		manager:   manager,
		closeChan: make(chan int),
		id:        atomic.AddUint64(&globalSessionId, 1),
	}
	if sendChanSize > 0 {
		session.sendChan = make(chan interface{}, sendChanSize)
		go session.sendLoop()
	}
	return session
}

func (session *Session) ID() uint64 {
	return session.id
}

func (session *Session) Codec() Codec {
	return session.codec
}

func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) == 1
}

func (session *Session) Close() error {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		err := session.codec.Close()
		close(session.closeChan)
		if session.manager != nil {
			session.manager.delSession(session)
		}
		session.invokeCloseCallback()
		return err

	}

	return SessionClosedError
}

func (session *Session) sendLoop() {
	defer session.Close()
	for {
		select {
		case msg := <-session.sendChan:
			if session.codec.Send(msg) != nil {
				return
			}

		case <-session.closeChan:
			return

		}
	}
}

func (session *Session) Recv() ([]byte, error) {
	msg, err := session.codec.Receive()
	if err != nil {
		session.Close()
	}
	return msg, err
}

func (session *Session) Send(msg interface{}) error {
	if session.IsClosed() {
		return SessionClosedError
	}
	if session.sendChan == nil {
		return session.codec.Send(msg)
	}
	select {
	case session.sendChan <- msg:
		return nil
	default:
		return SessionBlockedError
	}
}

func (session *Session) setReadDeadLine(t time.Duration) {
	if session.IsClosed() {
		return
	}

}

type CloseCallback struct {
	Handler interface{}
	Func    func()
}

func (session *Session) AddCloseCallback(handler interface{}, callback func()) {
	if session.IsClosed() {
		return
	}
	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()

	if session.closeCallbacks == nil {
		session.closeCallbacks = list.New()
	}
	c := CloseCallback{
		Handler: handler,
		Func:    callback,
	}
	session.closeCallbacks.PushBack(c)
}

func (session *Session) RemoveCloseCallback(handler interface{}) {
	if session.IsClosed() {
		return
	}
	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()

	for i := session.closeCallbacks.Front(); i != nil; i = i.Next() {
		if i.Value.(CloseCallback).Handler == handler {
			session.closeCallbacks.Remove(i)
			return
		}
	}
}

func (session *Session) invokeCloseCallback() {
	if session.IsClosed() {
		return
	}
	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()
	for i := session.closeCallbacks.Front(); i != nil; i = i.Next() {
		c := i.Value.(CloseCallback)
		c.Func()
	}
}
