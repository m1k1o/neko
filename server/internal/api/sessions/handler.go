package sessions

import (
	"github.com/m1k1o/neko/server/pkg/auth"
	"github.com/m1k1o/neko/server/pkg/types"
)

type SessionsHandler struct {
	sessions types.SessionManager
}

func New(
	sessions types.SessionManager,
) *SessionsHandler {
	// Init

	return &SessionsHandler{
		sessions: sessions,
	}
}

func (h *SessionsHandler) Route(r types.Router) {
	r.Get("/", h.sessionsList)

	r.With(auth.AdminsOnly).Route("/{sessionId}", func(r types.Router) {
		r.Get("/", h.sessionsRead)
		r.Delete("/", h.sessionsDelete)
		r.Post("/disconnect", h.sessionsDisconnect)
	})
}
