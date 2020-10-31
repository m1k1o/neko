package room

import (
	"github.com/go-chi/chi"

	"demodesk/neko/internal/api/utils"
	"demodesk/neko/internal/types"
)

type RoomHandler struct {
	sessions   types.SessionManager
	remote     types.RemoteManager
	broadcast  types.BroadcastManager
	websocket  types.WebSocketHandler
}

func New(
	sessions types.SessionManager,
	remote types.RemoteManager,
	broadcast types.BroadcastManager,
	websocket types.WebSocketHandler,
) *RoomHandler {
	// Init

	return &RoomHandler{
		sessions:   sessions,
		remote:     remote,
		broadcast:  broadcast,
		websocket:  websocket,
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
