package http

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (s *Server) debugHandlers() http.Handler {
	h := chi.NewMux()
	h.Route("/", func(r chi.Router) {
		h.Use(s.CheckProfiling())

		h.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, r.RequestURI+"/pprof/", http.StatusMovedPermanently)
		})
		h.HandleFunc("/pprof", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, r.RequestURI+"/", http.StatusMovedPermanently)
		})

		h.HandleFunc("/pprof/*", pprof.Index)
		h.HandleFunc("/pprof/cmdline", pprof.Cmdline)
		h.HandleFunc("/pprof/profile", pprof.Profile)
		h.HandleFunc("/pprof/symbol", pprof.Symbol)
		h.HandleFunc("/pprof/trace", pprof.Trace)
	})
	return h
}

func (s *Server) toggleDebugHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c := s.config

	c.Server.Profiling = !c.Server.Profiling
	w.Write([]byte(fmt.Sprintf("{\"Profiling\": %s}", strconv.FormatBool(c.Server.Profiling))))
}
