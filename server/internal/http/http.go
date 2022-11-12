package http

import (
	"context"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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
	router.Use(middleware.RequestLogger(&logformatter{logger}))
	router.Use(middleware.Recoverer) // Recover from panics without crashing server
	router.Use(middleware.Compress(5, "application/octet-stream"))

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

		path := webSocketHandler.MakeFilePath(filename)
		f, err := os.Open(path)
		if err != nil {
			http.Error(w, "not found or unable to open", http.StatusNotFound)
			return
		}
		defer f.Close()
		fileinfo, err := f.Stat()
		if err != nil {
			http.Error(w, "unable to stat file", http.StatusInternalServerError)
			return
		}

		buffer := make([]byte, fileinfo.Size())
		_, err = f.Read(buffer)
		if err != nil {
			http.Error(w, "error reading file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Write(buffer)
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

		r.ParseMultipartForm(32 << 20)
		buffer := make([]byte, FILE_UPLOAD_BUF_SIZE)
		for _, formheader := range r.MultipartForm.File["files"] {
			formfile, err := formheader.Open()
			if err != nil {
				logger.Warn().Err(err).Msg("failed to open formdata file")
				http.Error(w, "error writing file", http.StatusInternalServerError)
				return
			}
			f, err := os.OpenFile(webSocketHandler.MakeFilePath(formheader.Filename), os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				http.Error(w, "unable to open file for writing", http.StatusInternalServerError)
				return
			}

			var copied int64 = 0
			for copied < formheader.Size {
				var limit int64 = int64(len(buffer))
				if limit > formheader.Size-copied {
					limit = formheader.Size - copied
				}
				bytesRead, err := formfile.ReadAt(buffer[:limit], copied)
				if err != nil {
					logger.Warn().Err(err).Msg("failed copying file in upload")
					http.Error(w, "error writing file", http.StatusInternalServerError)
					return
				}
				f.Write(buffer[:bytesRead])
				copied += int64(bytesRead)
			}

			formfile.Close()
			f.Close()
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
