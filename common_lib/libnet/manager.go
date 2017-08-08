package libnet

import (
	"sync"
)

const sessionMapNum = 32

type sessionMap struct {
	sync.RWMutex
	sessions map[uint64]*Session
}

type Manager struct {
	sessionMaps [sessionMapNum]sessionMap
	disposeFlag bool
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
}

func NewManager() *Manager {
	manager := &Manager{}
	//初始化map
	for i := 0; i < len(manager.sessionMaps); i++ {
		manager.sessionMaps[i].sessions = make(map[uint64]*Session)
	}
	manager.disposeFlag = false

	return manager
}

func (m *Manager) Dispose() {
	m.disposeOnce.Do(func() {
		m.disposeFlag = true
		for i := 0; i < sessionMapNum; i++ {
			smap := &m.sessionMaps[i]
			smap.Lock()
			for _, session := range smap.sessions {
				session.Close()
			}
			smap.Unlock()
		}
		m.disposeWait.Wait()
	})
}

func (m *Manager) NewSession(codec Codec, sendChanSize int) *Session {
	session := newSession(m, codec, sendChanSize)
	m.putSession(session)
	return session
}

func (m *Manager) GetSession(sessionID uint64) *Session {
	smap := &m.sessionMaps[sessionID%sessionMapNum]
	smap.RLock()
	defer smap.RUnlock()
	session, _ := smap.sessions[sessionID]
	return session
}

func (m *Manager) putSession(session *Session) {
	smap := &m.sessionMaps[session.id%sessionMapNum]
	smap.Lock()
	defer smap.Unlock()
	smap.sessions[session.id] = session
	m.disposeWait.Add(1)
}

func (m *Manager) delSession(session *Session) {
	if m.disposeFlag {
		m.disposeWait.Done()
		return
	}
	smap := &m.sessionMaps[session.id%sessionMapNum]
	smap.Lock()
	defer smap.Unlock()
	delete(smap.sessions, session.id)
	m.disposeWait.Done()
}
