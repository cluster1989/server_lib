package perf

import (
	"net/http"
	"net/http/pprof"

	"github.com/thinkboy/log4go"
)

func Init(profBind []string) {
	pprofServeMux := http.NewServeMux()
	pprofServeMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	for _, addr := range profBind {
		go func() {
			if err := http.ListenAndServe(addr, pprofServeMux); err != nil {
				log4go.Error("http.ListenAndServe(\"%s\", pprofServeMux) error(%v)", addr, err)
				panic(err)
			}
		}()
	}
}
