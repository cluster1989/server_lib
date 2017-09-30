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

	//去注册
	go etcd.Register("/abdib/kal", "127.0.0.1:1234", time.Duration(10)*time.Second, time.Duration(5)*time.Second)
	go etcd.Watcher("/abdib/kal", watchBack)

	signal.InitSignal()
}

func watchBack(action string, key, val []byte) {
	logs.Debug("etcdcallback:action[%s],key[%s],value[%s]", action, string(key), string(val))
}
