package api

import (
	"net/http"

	"github.com/go-chi/chi"
	
	// "demodesk/neko/internal/api/member"
	// "demodesk/neko/internal/api/room"
)

func Mount(router *chi.Mux) {
	// all member routes
	router.Mount("/member", MemberRoutes())

	// all room routes
	router.Mount("/room", RoomRoutes())
}

func MemberRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		//nolint
		w.Write([]byte("hello world"))
	})

	return router
}

func RoomRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		//nolint
		w.Write([]byte("hello world"))
	})

	return router
}
