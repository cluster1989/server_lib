package etcd

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/wuqifei/server_lib/logs"
)

// 注册的etcd 服务
type EtcdService struct {
}

var (
	stopSignal = make(chan bool, 1)
)

// 注册自己到etcd服务
func Register(serviceKey string, val string, interval time.Duration, ttl time.Duration) {
	//因为有ttl机制，所以需要设置ticker，来保证注册
	if serviceKey[0] != '/' {
		serviceKey = "/" + serviceKey
	}
	go register(serviceKey, val, interval, ttl)
}

func register(serviceKey string, val string, interval, ttl time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		//先查询这个key是否存在
		err, vals := Get(serviceKey, false)
		if err != nil {

			//申请租约
			resp, err := etcdClient.Grant(context.Background(), int64(ttl/time.Second))
			if err != nil {
				logs.Error("etcd:grant failed[%v],key[%s],resp[%q]", err, serviceKey, resp)
			}
			//创建key-value 进去
			putResp, err := etcdClient.Put(context.Background(), serviceKey, val, clientv3.WithLease(resp.ID))
			if err != nil {
				logs.Error("etcd:put failed[%v],key[%s],resp[%q]", err, serviceKey, putResp)
			}
		} else {
			if len(vals) == 0 {
				//申请租约
				resp, err := etcdClient.Grant(context.Background(), int64(ttl/time.Second))
				if err != nil {
					logs.Error("etcd:grant failed[%v],key[%s],resp[%q]", err, serviceKey, resp)
				}
				//创建key-value 进去
				putResp, err := etcdClient.Put(context.Background(), serviceKey, val, clientv3.WithLease(resp.ID))
				if err != nil {
					logs.Error("etcd:put failed[%v],key[%s],resp[%q]", err, serviceKey, putResp)
				}
			} else {
				logs.Debug("etcd:registed key[%s] val[%v]", serviceKey, vals)
			}
		}
		//这里
		select {
		case <-stopSignal:
			return
		case <-ticker.C:
			//不做任何处理，一直循环
		}
	}
}

func Unregister(serviceKey string) {

	//停止这个注册服务
	stopSignal <- true
	//重置
	stopSignal = make(chan bool, 1)
	resp, err := etcdClient.Delete(context.Background(), serviceKey)
	if err != nil {
		logs.Error("etcd:delete failed [%v],key[%s] resp[%q]", err, serviceKey, resp)
	}
}
