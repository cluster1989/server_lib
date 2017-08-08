package zone_client

import (
	"sync"
)

const ClientMapNum = 32

var (
	Mananger *ClientManager
)

type ClientMap struct {
	sync.RWMutex
	Clients map[uint64]*Client
}

type ClientManager struct {
	ClientMaps  [ClientMapNum]ClientMap
	disposeFlag bool
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
}

func init() {
	Mananger = NewManager()
}

func NewManager() *ClientManager {
	manager := &ClientManager{}
	//初始化map
	for i := 0; i < len(manager.ClientMaps); i++ {
		manager.ClientMaps[i].Clients = make(map[uint64]*Client)
	}
	manager.disposeFlag = false
	return manager
}

func (m *ClientManager) Existed(clientId uint64) bool {
	smap := &m.ClientMaps[clientId%ClientMapNum]
	smap.RLock()
	defer smap.RUnlock()

	_, ok := smap.Clients[clientId]
	return ok
}

func (m *ClientManager) Dispose() {
	m.disposeOnce.Do(func() {
		m.disposeFlag = true
		for i := 0; i < ClientMapNum; i++ {
			smap := &m.ClientMaps[i]
			smap.Lock()
			for _, client := range smap.Clients {
				client.Close()
			}
			smap.Unlock()
		}
		m.disposeWait.Wait()
	})
}

func (m *ClientManager) GetClient(sessionId uint64) *Client {
	smap := &m.ClientMaps[sessionId%ClientMapNum]
	smap.RLock()
	defer smap.RUnlock()
	client, _ := smap.Clients[sessionId]
	return client
}

func (m *ClientManager) PutClient(client *Client) {
	sessionId := client.Session.ID()
	smap := &m.ClientMaps[sessionId%ClientMapNum]
	smap.Lock()
	defer smap.Unlock()
	smap.Clients[sessionId] = client
	m.disposeWait.Add(1)
}

func (m *ClientManager) DelClient(client *Client) {
	if m.disposeFlag {
		m.disposeWait.Done()
		return
	}
	sessionId := client.Session.ID()
	smap := &m.ClientMaps[sessionId%ClientMapNum]
	smap.Lock()
	defer smap.Unlock()
	delete(smap.Clients, sessionId)
	m.disposeWait.Done()
}
