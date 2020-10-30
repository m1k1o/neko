package member

import (
	"github.com/go-chi/chi"

	"demodesk/neko/internal/types"
)

type MemberHandler struct {
	sessions   types.SessionManager
	websocket  types.WebSocketHandler
}

func New(
	sessions types.SessionManager,
	websocket types.WebSocketHandler,
) *MemberHandler {
	// Init

	return &MemberHandler{
		sessions:   sessions,
		websocket:  websocket,
	}
}

func (h *MemberHandler) Router() *chi.Mux {
	r := chi.NewRouter()

	// TODO

	return r
}
