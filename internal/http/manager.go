package http

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/config"
	"demodesk/neko/internal/utils"
)

type HttpManagerCtx struct {
	logger zerolog.Logger
	router *chi.Mux
	http   *http.Server
	conf   *config.Server
}

func New(WebSocketManager types.WebSocketManager, ApiManager types.ApiManager, conf *config.Server) *HttpManagerCtx {
	logger := log.With().Str("module", "http").Logger()

	router := chi.NewRouter()
	router.Use(middleware.Recoverer) // Recover from panics without crashing server
	router.Use(cors.Handler(cors.Options{
		AllowOriginFunc:  func(r *http.Request, origin string) bool {
			return conf.AllowOrigin(origin)
		},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	router.Use(middleware.RequestID) // Create a request ID for each request
	router.Use(Logger) // Log API request calls using custom logger function

	router.Route("/api", ApiManager.Route)

	router.Get("/api/ws", func(w http.ResponseWriter, r *http.Request) {
		//nolint
		WebSocketManager.Upgrade(w, r, func(r *http.Request) bool {
			return conf.AllowOrigin(r.Header.Get("Origin"))
		})
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

	router.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.HttpNotFound(w)
	}))

	http := &http.Server{
		Addr:    conf.Bind,
		Handler: router,
	}

	return &HttpManagerCtx{
		logger: logger,
		router: router,
		http:   http,
		conf:   conf,
	}
}

func (s *HttpManagerCtx) Start() {
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
		s.logger.Info().Msgf("http listening on %s", s.http.Addr)
	}
}

func (s *HttpManagerCtx) Shutdown() error {
	return s.http.Shutdown(context.Background())
}
