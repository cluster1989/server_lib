package libkafka2

import (
	"fmt"

	"github.com/wuqifei/server_lib/libzookeeper"
)

var (
	ZKnotInitializedErr = fmt.Errorf("zookeeper not initialize")
)

//初始化zookeeper
func NewZK(option *libzookeeper.Option) (*libzookeeper.ZooKeeper, error) {

	return libzookeeper.NewZK(option)
}

func watchBrokers(zk *libzookeeper.ZooKeeper, callback chan<- []string) {

	watcher := make(chan []string, 0)
	zk.WatchChildren("brokers/ids", watcher)
	select {
	case <-watcher:
		brokers, err := fetchBrokers(zk)
		if err != nil {
			// 如果fetch 出错，zookeeper出问题
			panic(err)
		}
		callback <- brokers
	}
}

// 获取所有的kafka的broker
func fetchBrokers(zk *libzookeeper.ZooKeeper) ([]string, error) {

	return zk.KafkaBrokerList()
}
