package libzookeeper_test

import (
	"net"
	"testing"
	"time"

	"github.com/wuqifei/server_lib/libzookeeper"
)

var (
	zkServers = []string{"127.0.0.1:2181", "127.0.0.1:2182", "127.0.0.1:2183"}
)

func TestWatcher(t *testing.T) {
	option := &libzookeeper.Option{}
	option.Addrs = zkServers
	option.Timeout = time.Duration(2) * time.Second
	zk, _ := libzookeeper.NewZK(option)
	watcher := make(chan []string, 0)
	zk.WatchChildren("brokers/ids", watcher)

}

func TestBrokers(t *testing.T) {
	option := &libzookeeper.Option{}
	option.Addrs = zkServers
	option.Timeout = time.Duration(2) * time.Second
	zk, err := libzookeeper.NewZK(option)
	if err != nil {
		t.Fatal(err)
	}
	brokers, err := zk.KafkaBrokers()
	if err != nil {
		t.Fatal(err)

	}

	if len(brokers) == 0 {
		t.Error("no kafka")
	}

	for id, addr := range brokers {
		if conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond); err != nil {
			t.Logf("Failed to connect to Kafka broker %d at %s", id, addr)
		} else {
			conn.Close()
		}
	}
}

func TestBrokerList(t *testing.T) {
	option := &libzookeeper.Option{}
	option.Addrs = zkServers
	option.Timeout = time.Duration(2) * time.Second
	zk, err := libzookeeper.NewZK(option)
	if err != nil {
		t.Fatal(err)
	}
	brokers, err := zk.KafkaBrokerList()
	if err != nil {
		t.Fatal(err)

	}

	if len(brokers) == 0 {
		t.Error("no kafka")
	}

	for _, addr := range brokers {
		t.Logf("addr:[%s]", addr)
		if conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond); err != nil {
			t.Logf("Failed to connect to Kafka broker at %s", addr)
		} else {
			conn.Close()
		}
	}
}

func TestConsumergroups(t *testing.T) {

	option := &libzookeeper.Option{}
	option.Addrs = zkServers
	option.Timeout = time.Duration(2) * time.Second
	zkk, err := libzookeeper.NewZK(option)
	if err != nil {
		t.Fatal(err)
	}

	cg := zkk.NewConsumerGroup("abbgasobgasubgusabgsdngionisonignoas")

	if err := cg.Create(); err != nil {
		t.Fatal(err)
	}

	cgi := cg.NewInstance()
	t.Logf("cgi.id[%s]", cgi.ID)

	if err := cgi.Register([]string{"topic"}); err != nil {
		t.Fatal(err)
	}

	if err := cgi.Deregister(); err != nil {
		t.Fatal(err)
	}

	if err := cg.Delete(); err != nil {
		t.Fatal(err)
	}
}
