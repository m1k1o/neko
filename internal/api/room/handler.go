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

func (h *RoomHandler) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Route("/resolution", func(r chi.Router) {
		r.Get("/", h.ResolutionGet)
		r.Post("/", h.ResolutionChange)

		r.Get("/list", h.ResolutionList)
	})

	// TODO

	return r
}
