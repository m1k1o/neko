package http

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi"

	"github.com/demodesk/neko/pkg/types"
)

func pprofHandler(r types.Router) {
	r.Get("/debug/pprof/", func(w http.ResponseWriter, r *http.Request) error {
		pprof.Index(w, r)
		return nil
	})

	r.Get("/debug/pprof/{action}", func(w http.ResponseWriter, r *http.Request) error {
		action := chi.URLParam(r, "action")

		switch action {
		case "cmdline":
			pprof.Cmdline(w, r)
		case "profile":
			pprof.Profile(w, r)
		case "symbol":
			pprof.Symbol(w, r)
		case "trace":
			pprof.Trace(w, r)
		default:
			pprof.Handler(action).ServeHTTP(w, r)
		}

		return nil
	})
}
