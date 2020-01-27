package http

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"n.eko.moe/neko/internal/http/endpoint"
	"n.eko.moe/neko/internal/http/middleware"
	"n.eko.moe/neko/internal/types"
	"n.eko.moe/neko/internal/types/config"
)

type Server struct {
	logger zerolog.Logger
	router *chi.Mux
	http   *http.Server
	conf   *config.Server
}

func New(conf *config.Server, webSocketHandler types.WebSocketHandler) *Server {
	logger := log.With().Str("module", "webrtc").Logger()

	router := chi.NewRouter()
	// router.Use(middleware.Recoverer) // Recover from panics without crashing server
	router.Use(middleware.RequestID) // Create a request ID for each request
	router.Use(middleware.Logger)    // Log API request calls

	router.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		webSocketHandler.Upgrade(w, r)
	})

	fs := http.FileServer(http.Dir(conf.Static))
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(conf.Static + r.RequestURI); os.IsNotExist(err) {
			http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
		} else {
			fs.ServeHTTP(w, r)
		}
	})

	router.NotFound(endpoint.Handle(func(w http.ResponseWriter, r *http.Request) error {
		return &endpoint.HandlerError{
			Status:  http.StatusNotFound,
			Message: fmt.Sprintf("file '%s' is not found", r.RequestURI),
		}
	}))

	server := &http.Server{
		Addr:    conf.Bind,
		Handler: router,
	}

	return &Server{
		logger: logger,
		router: router,
		http:   server,
		conf:   conf,
	}
}

func (s *Server) Start() {
	if s.conf.Cert != "" && s.conf.Key != "" {
		go func() {
			if err := s.http.ListenAndServeTLS(s.conf.Cert, s.conf.Key); err != http.ErrServerClosed {
				s.logger.Panic().Err(err).Msg("unable to start https server")
			}
		}()
		s.logger.Info().Msgf("https listening on %s", s.http.Addr)
	} else {
		go func() {
			if err := s.http.ListenAndServe(); err != http.ErrServerClosed {
				s.logger.Panic().Err(err).Msg("unable to start http server")
			}
		}()
		s.logger.Warn().Msgf("http listening on %s", s.http.Addr)
	}
}

func (s *Server) Shutdown() error {
	return s.http.Shutdown(context.Background())
}
