package room

import (
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/http/auth"
)

type RoomHandler struct {
	logger    zerolog.Logger
	sessions  types.SessionManager
	desktop   types.DesktopManager
	capture   types.CaptureManager
}

func New(
	sessions types.SessionManager,
	desktop types.DesktopManager,
	capture types.CaptureManager,
) *RoomHandler {
	logger := log.With().
		Str("module", "api").
		Str("submodule", "room").
		Logger()

	return &RoomHandler{
		logger:    logger,
		sessions:  sessions,
		desktop:   desktop,
		capture:   capture,
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
	})

	r.Route("/screen", func(r chi.Router) {
		r.Get("/", h.screenConfiguration)

		r.With(auth.AdminsOnly).Post("/", h.screenConfigurationChange)
		r.With(auth.AdminsOnly).Get("/configurations", h.screenConfigurationsList)
	})
}
