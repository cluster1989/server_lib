package concurrent

import (
	"fmt"
	"io"
	"sync"
)

const SyncIDMapNum = 32

type SyncIdMap struct {
	sync.RWMutex
	Items map[uint64]io.Closer
}

type SyncGroupMap struct {
	SyncMaps    [SyncIDMapNum]SyncIdMap
	disposeFlag bool
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
}

// 新建一个map
func NewGroup() *SyncGroupMap {
	group := &SyncGroupMap{}
	for i := 0; i < len(group.SyncMaps); i++ {
		group.SyncMaps[i].Items = make(map[uint64]io.Closer)
	}
	group.disposeFlag = false
	return group
}

//释放，只执行一次
func (g *SyncGroupMap) Dispose() {
	g.disposeOnce.Do(func() {
		g.disposeFlag = true
		for i := 0; i < SyncIDMapNum; i++ {
			syncIDMap := &g.SyncMaps[i]
			syncIDMap.Lock()
			for key, item := range syncIDMap.Items {
				err := item.Close()
				//从group中删除
				delete(syncIDMap.Items, key)
				if err != nil {
					fmt.Printf("dispose sync map error:%d", key)
				}
			}

			syncIDMap.Unlock()
		}
		// 执行阻塞，直到所有都释放了
		g.disposeWait.Wait()
	})
}

func (g *SyncGroupMap) Get(id uint64) io.Closer {
	syncIDMap := g.SyncMaps[id%SyncIDMapNum]
	syncIDMap.Lock()
	defer syncIDMap.Unlock()
	item, _ := syncIDMap.Items[id]
	return item
}

func (g *SyncGroupMap) Set(id uint64, item io.Closer) {
	syncIDMap := g.SyncMaps[id%SyncIDMapNum]
	syncIDMap.Lock()
	defer syncIDMap.Unlock()
	syncIDMap.Items[id] = item
	g.disposeWait.Add(1)
}

func (g *SyncGroupMap) Del(id uint64) {
	if g.disposeFlag {
		g.disposeWait.Done()
		return
	}
	syncIDMap := g.SyncMaps[id%SyncIDMapNum]
	syncIDMap.Lock()
	defer syncIDMap.Unlock()
	delete(syncIDMap.Items, id)
	//-1
	g.disposeWait.Done()
}
