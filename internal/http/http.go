package http

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/config"
	"demodesk/neko/internal/http/endpoint"
)

type ServerCtx struct {
	logger zerolog.Logger
	router *chi.Mux
	http   *http.Server
	conf   *config.Server
}

func New(WebSocketManager types.WebSocketManager, ApiManager types.ApiManager, conf *config.Server) *ServerCtx {
	logger := log.With().Str("module", "http").Logger()

	router := chi.NewRouter()
	router.Use(middleware.Recoverer) // Recover from panics without crashing server
	router.Use(middleware.RequestID) // Create a request ID for each request
	router.Use(Logger) // Log API request calls using custom logger function

	ApiManager.Mount(router)

	router.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		if WebSocketManager.Upgrade(w, r) != nil {
			//nolint
			w.Write([]byte("unable to upgrade your connection to a websocket"))
		}
	})

	if conf.Static != "" {
		fs := http.FileServer(http.Dir(conf.Static))
		router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			if _, err := os.Stat(conf.Static + r.RequestURI); os.IsNotExist(err) {
				http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
			} else {
				fs.ServeHTTP(w, r)
			}
		})
	}

	router.NotFound(endpoint.Handle(func(w http.ResponseWriter, r *http.Request) error {
		return &endpoint.HandlerError{
			Status:  http.StatusNotFound,
			Message: fmt.Sprintf("Endpoint '%s' was not found.", r.RequestURI),
		}
	}))

	http := &http.Server{
		Addr:    conf.Bind,
		Handler: router,
	}

	return &ServerCtx{
		logger: logger,
		router: router,
		http:   http,
		conf:   conf,
	}
}

func (s *ServerCtx) Start() {
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

func (s *ServerCtx) Shutdown() error {
	return s.http.Shutdown(context.Background())
}
