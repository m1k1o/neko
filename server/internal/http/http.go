package http

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/config"
)

type Server struct {
	logger zerolog.Logger
	router *chi.Mux
	http   *http.Server
	conf   *config.Server
}

func New(conf *config.Server, webSocketHandler types.WebSocketHandler) *Server {
	logger := log.With().Str("module", "http").Logger()

	router := chi.NewRouter()
	router.Use(middleware.RequestID) // Create a request ID for each request
	router.Use(middleware.RequestLogger(&logformatter{logger}))
	router.Use(middleware.Recoverer) // Recover from panics without crashing server

	router.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		err := webSocketHandler.Upgrade(w, r)
		if err != nil {
			logger.Warn().Err(err).Msg("failed to upgrade websocket conection")
		}
	})

	router.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		password := r.URL.Query().Get("pwd")
		isAdmin, err := webSocketHandler.IsAdmin(password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if !isAdmin {
			http.Error(w, "bad authorization", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		stats := webSocketHandler.Stats()
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			logger.Warn().Err(err).Msg("failed writing json error response")
		}
	})

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("true"))
	})

	fs := http.FileServer(http.Dir(conf.Static))
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(conf.Static + r.URL.Path); !os.IsNotExist(err) {
			fs.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

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
