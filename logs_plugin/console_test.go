package logs_plugin_test

import (
	"testing"

	"github.com/wuqifei/server_lib/logs2"
	_ "github.com/wuqifei/server_lib/logs_plugin"
)

func TestConsole(t *testing.T) {
	logs2.Debug("12341512")
}
