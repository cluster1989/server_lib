package etcd

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
)

type Options struct {
	Tls         *tls.Config
	Endpoints   []string
	DialTimeout time.Duration
	Auth        *AuthOptions
}

type AuthOptions struct {
	UserName string
	Password string
}

var (
	etcdClient *clientv3.Client
	authClient clientv3.Auth
)

func NewOption() *Options {
	option := &Options{}
	option.DialTimeout = time.Duration(10) * time.Second
	return option
}

/**
 * 生成etcd 实例，在一个实例服务器中，有且只有一个etcd client实例
 */
func NewEtcd(options *Options) error {
	if etcdClient != nil {
		return nil
	}
	conf := clientv3.Config{}
	conf.Endpoints = options.Endpoints
	conf.DialTimeout = options.DialTimeout
	conf.TLS = options.Tls
	if options.Auth != nil {
		conf.Username = options.Auth.UserName
		conf.Password = options.Auth.Password
	}
	conf.Context = context.Background()

	client, err := clientv3.New(conf)
	if err != nil {
		fmt.Printf("ETCD:create etcd client failed,error(%v)\n", err)
		return err
	}

	if options.Auth != nil {
		authClient = clientv3.NewAuth(client)
	}

	etcdClient = client

	return nil
}

func Close() error {
	return etcdClient.Close()
}
