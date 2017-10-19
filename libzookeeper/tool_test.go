package libzookeeper_test

import (
	"testing"

	"github.com/wuqifei/server_lib/libzookeeper"
)

func TestParseZKAddrString(t *testing.T) {
	var (
		nodes  []string
		chroot string
	)

	nodes, chroot = libzookeeper.ParseZKAddrString("zk:0800,zk：2181/test")

	if len(nodes) != 2 || nodes[0] != "zk:0800" || nodes[1] != "zk：2181" {
		t.Error("Parsed nodes incorrectly:", nodes)
	}
	if chroot != "/test" {
		t.Error("Parsed chroot incorrectly:", chroot)
	}
}
