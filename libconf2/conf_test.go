package libconf2_test

import (
	"encoding/json"
	"testing"

	"github.com/wuqifei/server_lib/libconf2"
)

type Conf struct {
	T Test `b:"test"`
	A ABC  `b:"abc"`
}

type Test struct {
	A int64  `b:"a"`
	B string `b:"b"`
	C string `b:"c"`
}

type ABC struct {
	A string   `b:"a"`
	B string   `b:"b"`
	C string   `b:"c"`
	D []string `b:"d:,"`
}

var (
	conf *libconf2.Config
)

func init() {
	file := "test.conf"
	conf = libconf2.New()
	libconf2.Tag = "b"
	if err := conf.Parse(file); err != nil {
		panic(err)
	}
}

func TestConf(t *testing.T) {
	c := &Conf{}
	if err := conf.Unmarshal(c); err != nil {
		t.Error(err)
	}
	b, _ := json.Marshal(c)
	f, _ := json.Marshal(c.A)
	t.Logf("conf:[%s][%s] ---%s", string(b), string(f), c.A.D)
}
