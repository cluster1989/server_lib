package logs_plugin_test

import (
	"testing"
	"time"

	"github.com/wuqifei/server_lib/logs2"
	"github.com/wuqifei/server_lib/logs_plugin"
)

func TestLogsMutifile(t *testing.T) {

	m := &logs_plugin.MultiFileLogWriter{}
	m.Separate = []string{"emergency", "critical", "error", "warning", "info", "debug"}
	m.FullLogWriter = &logs_plugin.FileLogWriter{
		Daily:      true,
		MaxDays:    7,
		Rotate:     true,
		RotatePerm: "0440",
		Level:      logs2.LogLevelDebug,
		Perm:       "0660",
		Filename:   "logs/test.log",
	}

	logs2.DefaultLogger().Register("mutifile", m, logs_plugin.NewFilesWriter())
	logs2.DefaultLogger().SetDefaultLevel(logs2.LogLevelDebug)
	logs2.DefaultLogger().Async(3, 100)

	logs2.Debug("test 12341")
	logs2.Debug("test 12342")
	logs2.Debug("test 12343")
	logs2.Debug("test 12344")

	time.Sleep(time.Second * time.Duration(20))
}
