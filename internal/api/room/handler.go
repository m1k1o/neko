package room

import (
	"github.com/go-chi/chi"

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
	usersOnly func(chi.Router, func(chi.Router)),
	adminsOnly func(chi.Router, func(chi.Router)),
) *chi.Mux {
	r := chi.NewRouter()
	
	usersOnly(r, func(r chi.Router) {
		r.Get("/screen", h.ScreenConfiguration)
	})

	adminsOnly(r, func(r chi.Router) {
		r.Post("/screen", h.ScreenConfigurationChange)
		r.Get("/screen/configurations", h.ScreenConfigurationsList)
	})

	return r
}
