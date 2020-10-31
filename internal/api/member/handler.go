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

func (h *MemberHandler) Router(
	usersOnly func(chi.Router, func(chi.Router)),
	adminsOnly func(chi.Router, func(chi.Router)),
) *chi.Mux {
	r := chi.NewRouter()

	usersOnly(r, func(r chi.Router) {
		
	})

	adminsOnly(r, func(r chi.Router) {
		
	})

	return r
}
