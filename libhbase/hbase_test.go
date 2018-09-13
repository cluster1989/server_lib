package libhbase

import (
	"testing"
)

func TestHbase(t *testing.T) {
	c := New("127.0.0.1:2181")

	val, _ := c.GetRow("aaa", "row1")
	for _, v := range val {
		t.Logf("value:[%s]\n", string(v))
	}
	val, _ = c.GetWithQualifier("aaa", "row1", "vvvv", []string{"sad", "a"})
	for _, v := range val {
		t.Logf("value2:[%s]\n", string(v))
	}
}
