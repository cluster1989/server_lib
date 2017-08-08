package signal

import (
	"os"
	"syscall"

	"github.com/thinkboy/log4go"
)

// InitSignal register signals handler.
func InitSignal() {
	c := make(chan os.Signal, 1)
	defer close(c)
	for {
		s := <-c
		log4go.Info("server get a signal %s", s.String())
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
