package signal

import (
	"os"
	"os/signal"

	"github.com/wuqifei/server_lib/logs"
)

// InitSignal register signals handler.
func InitSignal() {
	chanSig := make(chan os.Signal, 1)
	signal.Notify(chanSig, os.Interrupt, os.Kill)
	sig := <-chanSig
	logs.Emergency("recv kill %s", sig.String())
}
