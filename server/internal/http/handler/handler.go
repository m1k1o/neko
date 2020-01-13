package handler

import (
	"fmt"
	"net/http"
	"os"

	"n.eko.moe/neko/internal/http/middleware"
	"n.eko.moe/neko/internal/http/endpoint"
  "n.eko.moe/neko/internal/webrtc"

	"github.com/go-chi/chi"
)

type Handler struct {
	router  *chi.Mux
	manager *webrtc.WebRTCManager
}

func New(password, static string) *chi.Mux {
	router := chi.NewRouter()
	manager, err := webrtc.NewManager(password)
	if err != nil {
		panic(err)
	}

	handler := &Handler{
		router:  router,
		manager: manager,
	}

	router.Use(middleware.Recoverer) // Recover from panics without crashing server
	// router.Use(middleware.Logger)    // Log API request calls

	router.Get("/ping", endpoint.Handle(handler.Ping))
  router.Get("/ws", endpoint.Handle(handler.WebSocket))

	fs := http.FileServer(http.Dir(static))
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(static + r.RequestURI); os.IsNotExist(err) {
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

	return router
}
