package room

import (
	"github.com/go-chi/chi"

	"demodesk/neko/internal/types"
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
	// TODO: Authorizaton.
	r.Route("/screen", func(r chi.Router) {
		r.Get("/", h.ScreenConfiguration)
		r.Post("/", h.ScreenConfigurationChange)

		r.Get("/configurations", h.ScreenConfigurationsList)
	})

	r.Route("/clipboard", func(r chi.Router) {
		r.Get("/", h.ClipboardRead)
		r.Post("/", h.ClipboardWrite)
	})
}
