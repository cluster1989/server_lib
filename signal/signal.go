package signal

import (
	"fmt"
	"os"
	"syscall"
)

// InitSignal register signals handler.
func InitSignal() {
	c := make(chan os.Signal, 1)
	defer close(c)
	for {
		s := <-c
		fmt.Printf("server get a signal %s", s.String())
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
