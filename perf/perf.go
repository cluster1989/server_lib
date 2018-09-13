package perf

import (
	"net/http"
	"net/http/pprof"
)

func Init(profBind []string) {
	pprofServeMux := http.NewServeMux()
	pprofServeMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	for _, addr := range profBind {

		go func(profaddr string) {
			http.ListenAndServe(profaddr, pprofServeMux)
			if err := http.ListenAndServe(profaddr, pprofServeMux); err != nil {
				panic(err)
			}

		}(addr)
	}
}
