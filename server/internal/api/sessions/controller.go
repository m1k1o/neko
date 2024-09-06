package sessions

import (
	"errors"
	"net/http"

	"github.com/demodesk/neko/pkg/auth"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
	"github.com/go-chi/chi"
)

type SessionDataPayload struct {
	ID      string              `json:"id"`
	Profile types.MemberProfile `json:"profile"`
	State   types.SessionState  `json:"state"`
}

func (h *SessionsHandler) sessionsList(w http.ResponseWriter, r *http.Request) error {
	sessions := []SessionDataPayload{}
	for _, session := range h.sessions.List() {
		sessions = append(sessions, SessionDataPayload{
			ID:      session.ID(),
			Profile: session.Profile(),
			State:   session.State(),
		})
	}

	return utils.HttpSuccess(w, sessions)
}

func (h *SessionsHandler) sessionsRead(w http.ResponseWriter, r *http.Request) error {
	sessionId := chi.URLParam(r, "sessionId")

	session, ok := h.sessions.Get(sessionId)
	if !ok {
		return utils.HttpNotFound("session not found")
	}

	return utils.HttpSuccess(w, SessionDataPayload{
		ID:      session.ID(),
		Profile: session.Profile(),
		State:   session.State(),
	})
}

func (h *SessionsHandler) sessionsDelete(w http.ResponseWriter, r *http.Request) error {
	session, _ := auth.GetSession(r)

	sessionId := chi.URLParam(r, "sessionId")
	if sessionId == session.ID() {
		return utils.HttpBadRequest("cannot delete own session")
	}

	err := h.sessions.Delete(sessionId)
	if err != nil {
		if errors.Is(err, types.ErrSessionNotFound) {
			return utils.HttpBadRequest("session not found")
		} else {
			return utils.HttpInternalServerError().WithInternalErr(err)
		}
	}

	return utils.HttpSuccess(w)
}

func (h *SessionsHandler) sessionsDisconnect(w http.ResponseWriter, r *http.Request) error {
	sessionId := chi.URLParam(r, "sessionId")

	err := h.sessions.Disconnect(sessionId)
	if err != nil {
		if errors.Is(err, types.ErrSessionNotFound) {
			return utils.HttpBadRequest("session not found")
		} else {
			return utils.HttpInternalServerError().WithInternalErr(err)
		}
	}

	return utils.HttpSuccess(w)
}
