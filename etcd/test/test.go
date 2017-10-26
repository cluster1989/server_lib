package main

import (
	"time"

	"github.com/wuqifei/server_lib/etcd"
	"github.com/wuqifei/server_lib/logs"
	"github.com/wuqifei/server_lib/signal"
)

func main() {
	options := &etcd.Options{}
	options.Endpoints = []string{"127.0.0.1:2379"}
	options.DialTimeout = time.Duration(2) * time.Second

	etcd.NewEtcd(options)
	etcd.Set("/test/test1", "adn")

	e, v := etcd.Get("/test/test2", false)
	for _, b := range v {
		logs.Debug("etcd value:[%s] len[%d]", b, len(v))

	}
	logs.Debug("etcd get :value [%v]len [%d]", v, len(v))
	if e != nil {
		logs.Debug("etcd get value error:[%v] ", e)
	}
	//去注册
	go etcd.Register("/abdib/kal", "127.0.0.1:1234", time.Duration(10)*time.Second, time.Duration(5)*time.Second)
	go etcd.Watcher("/abdib/kal", watchBack)

	signal.InitSignal()
}

func watchBack(action string, key, val []byte) {
	logs.Debug("etcdcallback:action[%s],key[%s],value[%s]", action, string(key), string(val))
}
