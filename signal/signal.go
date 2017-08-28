package signal

import (
	"os"
	"syscall"

	"github.com/wuqifei/server_lib/logs"
)

// InitSignal register signals handler.
func InitSignal() {
	c := make(chan os.Signal, 1)
	defer close(c)
	for {
		s := <-c
		logs.Emergency("server get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			return
		case syscall.SIGHUP:
			continue
		default:
			return
		}
	}
}
