package http

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/internal/config"
	"github.com/demodesk/neko/pkg/types"
)

type HttpManagerCtx struct {
	logger zerolog.Logger
	config *config.Server
	router types.Router
	http   *http.Server
}

func New(WebSocketManager types.WebSocketManager, ApiManager types.ApiManager, config *config.Server) *HttpManagerCtx {
	logger := log.With().Str("module", "http").Logger()

	router := newRouter(logger)
	router.UseBypass(cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return config.AllowOrigin(origin)
		},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	router.Route("/api", ApiManager.Route)

	router.Get("/api/ws", WebSocketManager.Upgrade(func(r *http.Request) bool {
		return config.AllowOrigin(r.Header.Get("Origin"))
	}))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) error {
		_, err := w.Write([]byte("true"))
		return err
	})

	if config.Metrics {
		router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) error {
			promhttp.Handler().ServeHTTP(w, r)
			return nil
		})
	}

	if config.Static != "" {
		fs := http.FileServer(http.Dir(config.Static))
		router.Get("/*", func(w http.ResponseWriter, r *http.Request) error {
			_, err := os.Stat(config.Static + r.URL.Path)
			if err == nil {
				fs.ServeHTTP(w, r)
				return nil
			}
			if os.IsNotExist(err) {
				http.NotFound(w, r)
				return nil
			}
			return err
		})
	}

	if config.PProf {
		pprofHandler(router)
	}

	return &HttpManagerCtx{
		logger: logger,
		config: config,
		router: router,
		http: &http.Server{
			Addr:    config.Bind,
			Handler: router,
		},
	}
}

func (manager *HttpManagerCtx) Start() {
	if manager.config.Cert != "" && manager.config.Key != "" {
		go func() {
			if err := manager.http.ListenAndServeTLS(manager.config.Cert, manager.config.Key); err != http.ErrServerClosed {
				manager.logger.Panic().Err(err).Msg("unable to start https server")
			}
		}()
		manager.logger.Info().Msgf("https listening on %s", manager.http.Addr)
	} else {
		go func() {
			if err := manager.http.ListenAndServe(); err != http.ErrServerClosed {
				manager.logger.Panic().Err(err).Msg("unable to start http server")
			}
		}()
		manager.logger.Info().Msgf("http listening on %s", manager.http.Addr)
	}
}

func (manager *HttpManagerCtx) Shutdown() error {
	manager.logger.Info().Msg("shutdown")

	return manager.http.Shutdown(context.Background())
}
