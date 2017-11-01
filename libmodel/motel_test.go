package libmodel

import (
	"testing"
)

type AfbasniAiios struct {
}

func TestSQLStr(t *testing.T) {
	name := ObjName2SqlName("AfbasniAiios")
	if name != "afbasni_aiios" {
		t.Fatal("name error:[%s]", name)
	}
}
