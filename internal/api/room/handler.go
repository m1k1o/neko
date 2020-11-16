package room

import (
	"github.com/go-chi/chi"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/http/auth"
)

type RoomHandler struct {
	sessions  types.SessionManager
	desktop   types.DesktopManager
	capture  types.CaptureManager
}

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
) *RoomHandler {
	// Init

	return &RoomHandler{
		sessions:  sessions,
		desktop:   desktop,
		capture:   capture,
	}
}

func (h *RoomHandler) Route(r chi.Router) {
	r.Route("/screen", func(r chi.Router) {
		r.Get("/", h.ScreenConfiguration)

		r.With(auth.AdminsOnly).Post("/", h.ScreenConfigurationChange)
		r.With(auth.AdminsOnly).Get("/configurations", h.ScreenConfigurationsList)
	})

	r.With(auth.HostsOnly).Route("/clipboard", func(r chi.Router) {
		r.Get("/", h.ClipboardRead)
		r.Post("/", h.ClipboardWrite)
	})
}
