package room

import (
	"github.com/go-chi/chi"

	"demodesk/neko/internal/api/utils"
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

func (h *RoomHandler) Router(
	usersOnly utils.HttpMiddleware,
	adminsOnly utils.HttpMiddleware,
) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/screen", func(r chi.Router) {
		r.With(usersOnly).Get("/", h.ScreenConfiguration)
		r.With(adminsOnly).Post("/", h.ScreenConfigurationChange)

		r.With(adminsOnly).Get("/configurations", h.ScreenConfigurationsList)
	})

	r.With(adminsOnly).Route("/clipboard", func(r chi.Router) {
		r.Get("/", h.ClipboardRead)
		r.Post("/", h.ClipboardWrite)
	})

	return r
}
