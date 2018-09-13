package concurrent

import (
	"fmt"
	"io"
	"sync"
)

const ConcurrentMapNum = 32

type ConcurrentIDMap struct {
	sync.RWMutex
	Items map[uint64]interface{}
}

type ConcurrentIDGroupMap struct {
	SyncMaps    [ConcurrentMapNum]ConcurrentIDMap
	disposeFlag bool
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
	count       int32
}

// 新建一个map
func NewCocurrentIDGroup() *ConcurrentIDGroupMap {
	group := &ConcurrentIDGroupMap{}
	for i := 0; i < len(group.SyncMaps); i++ {
		group.SyncMaps[i].Items = make(map[uint64]interface{})
	}
	group.disposeFlag = false
	return group
}

//释放，只执行一次
func (g *ConcurrentIDGroupMap) Dispose() {
	g.disposeOnce.Do(func() {
		g.disposeFlag = true
		for i := 0; i < ConcurrentMapNum; i++ {
			syncIDMap := &g.SyncMaps[i]
			syncIDMap.Lock()
			for key, item := range syncIDMap.Items {

				delete(syncIDMap.Items, key)
				g.disposeWait.Done()
				var err error
				switch item.(type) {
				case io.Closer:
					closer := item.(io.Closer)
					err = closer.Close()
				default:
				}
				//从group中删除
				if err != nil {
					fmt.Printf("err :concurrent map :dispose map key:%d error:%v\n", key, err)
				}
			}

			syncIDMap.Unlock()
		}
		g.count = 0
		// 执行阻塞，直到所有都释放了
		g.disposeWait.Wait()
	})
}

func (g *ConcurrentIDGroupMap) Get(id uint64) interface{} {
	syncIDMap := g.SyncMaps[id%ConcurrentMapNum]
	syncIDMap.Lock()
	defer syncIDMap.Unlock()
	item, _ := syncIDMap.Items[id]
	return item
}

func (g *ConcurrentIDGroupMap) Set(id uint64, item interface{}) {
	syncIDMap := g.SyncMaps[id%ConcurrentMapNum]
	syncIDMap.Lock()
	defer syncIDMap.Unlock()
	syncIDMap.Items[id] = item
	g.disposeWait.Add(1)
	g.count++
}

func (g *ConcurrentIDGroupMap) Del(id uint64) {
	if g.disposeFlag {
		g.disposeWait.Done()
		return
	}
	syncIDMap := g.SyncMaps[id%ConcurrentMapNum]
	syncIDMap.Lock()
	defer syncIDMap.Unlock()
	delete(syncIDMap.Items, id)

	g.disposeWait.Done()
	g.count--
}

func (g *ConcurrentIDGroupMap) Count() int32 {
	return g.count
}
