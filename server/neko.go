package neko

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"

	"n.eko.moe/neko/internal/config"
	"n.eko.moe/neko/internal/http/endpoint"
	"n.eko.moe/neko/internal/http/middleware"
	"n.eko.moe/neko/internal/structs"
	"n.eko.moe/neko/internal/webrtc"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	//
	buildDate = ""
	//
	gitCommit = ""
	//
	gitVersion = ""
	//
	gitState = ""
	// Major version when you make incompatible API changes,
	major = "0"
	// Minor version when you add functionality in a backwards-compatible manner, and
	minor = "0"
	// Patch version when you make backwards-compatible bug fixeneko.
	patch = "0"
)

var Service *Neko

func init() {
	Service = &Neko{
		Version: &structs.Version{
			Major:        major,
			Minor:        minor,
			Patch:        patch,
			GitVersion:   gitVersion,
			GitCommit:    gitCommit,
			GitTreeState: gitState,
			BuildDate:    buildDate,
			GoVersion:    runtime.Version(),
			Compiler:     runtime.Compiler,
			Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		},
		Root:  &config.Root{},
		Serve: &config.Serve{},
	}
}

type Neko struct {
	Version *structs.Version
	Root    *config.Root
	Serve   *config.Serve
	Logger  zerolog.Logger
	http    *http.Server
}

func (neko *Neko) Preflight() {
	neko.Logger = log.With().Str("service", "neko").Logger()
}

func (neko *Neko) Start() {
	router := chi.NewRouter()

	manager, err := webrtc.NewManager(neko.Serve.Password)
	if err != nil {
		neko.Logger.Panic().Err(err).Msg("Can not start webrtc manager")
	}

	router.Use(middleware.Recoverer) // Recover from panics without crashing server
	router.Use(middleware.RequestID) // Create a request ID for each request
	router.Use(middleware.Logger)    // Log API request calls

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("."))
	})

	router.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		if err := manager.Upgrade(w, r); err != nil {
			neko.Logger.Error().Err(err).Msg("session.destroy has failed")
		}
	})

	fs := http.FileServer(http.Dir(neko.Serve.Static))
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(neko.Serve.Static + r.RequestURI); os.IsNotExist(err) {
			http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
		} else {
			fs.ServeHTTP(w, r)
		}
	})

	router.NotFound(endpoint.Handle(func(w http.ResponseWriter, r *http.Request) error {
		return &endpoint.HandlerError{
			Status:  http.StatusNotFound,
			Message: fmt.Sprintf("Endpoint '%s' is not avalible", r.RequestURI),
		}
	}))

	server := &http.Server{
		Addr:    neko.Serve.Bind,
		Handler: router,
	}

	if neko.Serve.Cert != "" && neko.Serve.Key != "" {
		go func() {
			if err := server.ListenAndServeTLS(neko.Serve.Cert, neko.Serve.Key); err != http.ErrServerClosed {
				neko.Logger.Panic().Err(err).Msg("Unable to start https server")
			}
		}()
		neko.Logger.Info().Msgf("HTTPS listening on %s", server.Addr)
	} else {
		go func() {
			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				neko.Logger.Panic().Err(err).Msg("Unable to start http server")
			}
		}()
		neko.Logger.Warn().Msgf("HTTP listening on %s", server.Addr)
	}

	neko.http = server
}

func (neko *Neko) Shutdown() {
	if neko.http != nil {
		if err := neko.http.Shutdown(context.Background()); err != nil {
			neko.Logger.Err(err).Msg("HTTP server shutdown with an error")
		} else {
			neko.Logger.Debug().Msg("HTTP server shutdown")
		}
	}
}

func (neko *Neko) ServeCommand(cmd *cobra.Command, args []string) {
	neko.Logger.Info().Msg("Starting HTTP/S server")
	neko.Start()

	neko.Logger.Info().Msg("Service ready")

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit

	neko.Logger.Warn().Msgf("Received %s, attempting graceful shutdown: \n", sig)

	neko.Shutdown()
	neko.Logger.Info().Msg("Shutting down complete")
}
