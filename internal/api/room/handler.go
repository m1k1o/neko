package room

import (
	"context"
	"net/http"

	"demodesk/neko/internal/http/auth"
	"demodesk/neko/internal/types"
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

func (h *RoomHandler) Route(r types.Router) {
	r.With(auth.AdminsOnly).Route("/broadcast", func(r types.Router) {
		r.Get("/", h.broadcastStatus)
		r.Post("/start", h.boradcastStart)
		r.Post("/stop", h.boradcastStop)
	})

	r.With(auth.CanAccessClipboardOnly).With(auth.HostsOnly).Route("/clipboard", func(r types.Router) {
		r.Get("/", h.clipboardGetText)
		r.Post("/", h.clipboardSetText)
		r.Get("/image.png", h.clipboardGetImage)

		// TODO: Refactor. xclip is failing to set propper target type
		// and this content is sent back to client as text in another
		// clipboard update. Therefore endpoint is not usable!
		//r.Post("/image", h.clipboardSetImage)

		// TODO: Refactor. If there would be implemented custom target
		// retrieval, this endpoint would be useful.
		//r.Get("/targets", h.clipboardGetTargets)
	})

	r.With(auth.CanHostOnly).Route("/keyboard", func(r types.Router) {
		r.Get("/map", h.keyboardMapGet)
		r.With(auth.HostsOnly).Post("/map", h.keyboardMapSet)

		r.Get("/modifiers", h.keyboardModifiersGet)
		r.With(auth.HostsOnly).Post("/modifiers", h.keyboardModifiersSet)
	})

	r.With(auth.CanHostOnly).Route("/control", func(r types.Router) {
		r.Get("/", h.controlStatus)
		r.Post("/request", h.controlRequest)
		r.Post("/release", h.controlRelease)

		r.With(auth.AdminsOnly).Post("/take", h.controlTake)
		r.With(auth.AdminsOnly).Post("/give", h.controlGive)
		r.With(auth.AdminsOnly).Post("/reset", h.controlReset)
	})

	r.With(auth.CanWatchOnly).Route("/screen", func(r types.Router) {
		r.Get("/", h.screenConfiguration)
		r.With(auth.AdminsOnly).Post("/", h.screenConfigurationChange)
		r.With(auth.AdminsOnly).Get("/configurations", h.screenConfigurationsList)

		r.Get("/cast.jpg", h.screenCastGet)
		r.With(auth.AdminsOnly).Get("/shot.jpg", h.screenShotGet)
	})

	r.With(h.uploadMiddleware).Route("/upload", func(r types.Router) {
		r.Post("/drop", h.uploadDrop)
		r.Post("/dialog", h.uploadDialogPost)
		r.Delete("/dialog", h.uploadDialogClose)
	})
}

func (h *RoomHandler) uploadMiddleware(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	session, ok := auth.GetSession(r)
	if !ok || (!session.IsHost() && (!session.Profile().CanHost || !h.sessions.ImplicitHosting())) {
		return nil, utils.HttpForbidden("without implicit hosting, only host can upload files")
	}

	return nil, nil
}
