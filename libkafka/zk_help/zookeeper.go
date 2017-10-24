package zk_help

import (
	"fmt"

	"github.com/wuqifei/server_lib/libzookeeper"
)

var (
	ZKnotInitializedErr = fmt.Errorf("zookeeper not initialize")
)

var (
	Zookeeper *libzookeeper.ZooKeeper
)

//初始化zookeeper
func Init(option *libzookeeper.Option) error {
	if Zookeeper != nil {
		return nil
	}
	zk, err := libzookeeper.NewZK(option)
	Zookeeper = zk
	return err
}

// 获取所有的kafka的broker
func FetchBrokers() ([]string, error) {
	if Zookeeper == nil {
		return nil, ZKnotInitializedErr
	}

	return Zookeeper.KafkaBrokerList()
}

// 关闭zk
func Close() error {
	return Zookeeper.Close()
}
