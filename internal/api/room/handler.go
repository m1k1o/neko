package room

import (
	"net/http"

	"github.com/go-chi/chi"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/http/auth"
	"demodesk/neko/internal/utils"
)

type RoomHandler struct {
	sessions types.SessionManager
	desktop  types.DesktopManager
	capture  types.CaptureManager
}

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
) *RoomHandler {
	// Init

	return &RoomHandler{
		sessions: sessions,
		desktop:  desktop,
		capture:  capture,
	}
}

func (h *RoomHandler) Route(r chi.Router) {
	r.With(auth.AdminsOnly).Route("/broadcast", func(r chi.Router) {
		r.Get("/", h.broadcastStatus)
		r.Post("/start", h.boradcastStart)
		r.Post("/stop", h.boradcastStop)
	})

	r.With(auth.HostsOnly).Route("/clipboard", func(r chi.Router) {
		r.Get("/", h.clipboardRead)
		r.Post("/", h.clipboardWrite)
	})

	r.With(auth.HostsOnly).Route("/keyboard", func(r chi.Router) {
		r.Post("/layout", h.keyboardLayoutSet)
		r.Post("/modifiers", h.keyboardModifiersSet)
	})

	r.Route("/control", func(r chi.Router) {
		r.Get("/", h.controlStatus)
		r.Post("/request", h.controlRequest)
		r.Post("/release", h.controlRelease)

		r.With(auth.AdminsOnly).Post("/take", h.controlTake)
		r.With(auth.AdminsOnly).Post("/give", h.controlGive)
		r.With(auth.AdminsOnly).Post("/reset", h.controlReset)
	})

	r.Route("/screen", func(r chi.Router) {
		r.Get("/", h.screenConfiguration)

		r.With(auth.AdminsOnly).Post("/", h.screenConfigurationChange)
		r.With(auth.AdminsOnly).Get("/configurations", h.screenConfigurationsList)
	})

	r.With(h.uploadMiddleware).Route("/upload", func(r chi.Router) {
		r.Post("/drop", h.uploadDrop)
	})
}

func (h *RoomHandler) uploadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := auth.GetSession(r)
		if !session.IsHost() && !h.sessions.ImplicitHosting() {
			utils.HttpForbidden(w, "Without implicit hosting, only host can upload files.")
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
