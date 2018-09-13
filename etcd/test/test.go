package main

import (
	"fmt"
	"time"

	"github.com/wuqifei/server_lib/etcd"
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
		fmt.Printf("etcd value:[%s] len[%d]\n", b, len(v))

	}
	fmt.Printf("etcd get :value [%v]len [%d]\n", v, len(v))
	if e != nil {
		fmt.Printf("etcd get value error:[%v] \n", e)
	}
	//去注册
	go etcd.Register("/abdib/kal", "127.0.0.1:1234", time.Duration(10)*time.Second, time.Duration(5)*time.Second)
	go etcd.Watcher("/abdib/kal", watchBack)

	signal.InitSignal()
}

func watchBack(action string, key, val []byte) {
	fmt.Printf("etcdcallback:action[%s],key[%s],value[%s]\n", action, string(key), string(val))
}
