package http

import (
	"context"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/config"
	"m1k1o/neko/internal/types"
)

const FILE_UPLOAD_BUF_SIZE = 65000

type Server struct {
	logger zerolog.Logger
	router *chi.Mux
	http   *http.Server
	conf   *config.Server
}

func New(conf *config.Server, webSocketHandler types.WebSocketHandler, desktop types.DesktopManager) *Server {
	logger := log.With().Str("module", "http").Logger()

	router := chi.NewRouter()
	router.Use(middleware.RequestID) // Create a request ID for each request
	if conf.Proxy {
		router.Use(middleware.RealIP)
	}
	router.Use(middleware.RequestLogger(&logformatter{logger}))
	router.Use(middleware.Recoverer) // Recover from panics without crashing server
	router.Use(middleware.Compress(5, "application/octet-stream"))

	router.Use(cors.Handler(cors.Options{
		AllowOriginFunc:  conf.AllowOrigin,
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	if conf.PathPrefix != "/" {
		router.Use(func(h http.Handler) http.Handler {
			return http.StripPrefix(conf.PathPrefix, h)
		})
	}

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

	router.Get("/screenshot.jpg", func(w http.ResponseWriter, r *http.Request) {
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

		if webSocketHandler.IsLocked("login") {
			http.Error(w, "room is locked", http.StatusLocked)
			return
		}

		quality, err := strconv.Atoi(r.URL.Query().Get("quality"))
		if err != nil {
			quality = 90
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Content-Type", "image/jpeg")

		img := desktop.GetScreenshotImage()
		if err := jpeg.Encode(w, img, &jpeg.Options{Quality: quality}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// allow downloading and uploading files
	if webSocketHandler.FileTransferEnabled() {
		router.Get("/file", func(w http.ResponseWriter, r *http.Request) {
			password := r.URL.Query().Get("pwd")
			isAuthorized, err := webSocketHandler.CanTransferFiles(password)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			if !isAuthorized {
				http.Error(w, "bad authorization", http.StatusUnauthorized)
				return
			}

			filename := r.URL.Query().Get("filename")
			badChars, _ := regexp.MatchString(`(?m)\.\.(?:\/|$)`, filename)
			if filename == "" || badChars {
				http.Error(w, "bad filename", http.StatusBadRequest)
				return
			}

			filePath := webSocketHandler.FileTransferPath(filename)
			f, err := os.Open(filePath)
			if err != nil {
				http.Error(w, "not found or unable to open", http.StatusNotFound)
				return
			}
			defer f.Close()

			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
			io.Copy(w, f)
		})

		router.Post("/file", func(w http.ResponseWriter, r *http.Request) {
			password := r.URL.Query().Get("pwd")
			isAuthorized, err := webSocketHandler.CanTransferFiles(password)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			if !isAuthorized {
				http.Error(w, "bad authorization", http.StatusUnauthorized)
				return
			}

			err = r.ParseMultipartForm(32 << 20)
			if err != nil || r.MultipartForm == nil {
				logger.Warn().Err(err).Msg("failed to parse multipart form")
				http.Error(w, "error parsing form", http.StatusBadRequest)
				return
			}

			for _, formheader := range r.MultipartForm.File["files"] {
				filePath := webSocketHandler.FileTransferPath(formheader.Filename)

				formfile, err := formheader.Open()
				if err != nil {
					logger.Warn().Err(err).Msg("failed to open formdata file")
					http.Error(w, "error writing file", http.StatusInternalServerError)
					return
				}
				defer formfile.Close()

				f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
				if err != nil {
					http.Error(w, "unable to open file for writing", http.StatusInternalServerError)
					return
				}
				defer f.Close()

				io.Copy(f, formfile)
			}

			err = r.MultipartForm.RemoveAll()
			if err != nil {
				logger.Warn().Err(err).Msg("failed to remove multipart form")
			}
		})
	}

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
