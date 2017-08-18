package concurrent

import (
	"sync"
)

const MapNum = 32

type SyncUint64ItemMap struct {
	sync.RWMutex
	Items map[uint64]interface{}
}

type SyncUint64GroupMap struct {
	SyncMaps    [MapNum]SyncUint64ItemMap
	disposeFlag bool
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
}

// 新建一个map
func NewUint64Group() *SyncUint64GroupMap {
	group := &SyncUint64GroupMap{}
	for i := 0; i < len(group.SyncMaps); i++ {
		group.SyncMaps[i].Items = make(map[uint64]interface{})
	}
	group.disposeFlag = false
	return group
}

//释放，只执行一次
func (g *SyncUint64GroupMap) Dispose() {
	g.disposeOnce.Do(func() {
		g.disposeFlag = true
		for i := 0; i < MapNum; i++ {
			syncIDMap := &g.SyncMaps[i]
			syncIDMap.Lock()
			for key, _ := range syncIDMap.Items {
				//从group中删除
				delete(syncIDMap.Items, key)
			}

			syncIDMap.Unlock()
		}
		// 执行阻塞，直到所有都释放了
		g.disposeWait.Wait()
	})
}

func (g *SyncUint64GroupMap) Get(id uint64) interface{} {
	syncIDMap := g.SyncMaps[id%MapNum]
	syncIDMap.Lock()
	defer syncIDMap.Unlock()
	item, _ := syncIDMap.Items[id]
	return item
}

func (g *SyncUint64GroupMap) Set(id uint64, item interface{}) {
	syncIDMap := g.SyncMaps[id%MapNum]
	syncIDMap.Lock()
	defer syncIDMap.Unlock()
	syncIDMap.Items[id] = item
	g.disposeWait.Add(1)
}

func (g *SyncUint64GroupMap) Del(id uint64) {
	if g.disposeFlag {
		g.disposeWait.Done()
		return
	}
	syncIDMap := g.SyncMaps[id%MapNum]
	syncIDMap.Lock()
	defer syncIDMap.Unlock()
	delete(syncIDMap.Items, id)
	//-1
	g.disposeWait.Done()
}
