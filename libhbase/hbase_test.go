package libhbase

import (
	"testing"

	"github.com/wuqifei/server_lib/logs"
)

func TestHbase(t *testing.T) {
	c := New("127.0.0.1:2181")

	val, _ := c.GetRow("aaa", "row1")
	for _, v := range val {
		logs.Debug("value:[%s]", string(v))
	}
	val, _ = c.GetWithQualifier("aaa", "row1", "vvvv", []string{"sad", "a"})
	for _, v := range val {
		logs.Debug("value2:[%s]", string(v))
	}
}
