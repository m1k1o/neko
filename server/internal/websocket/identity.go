package websocket

import (
	"n.eko.moe/neko/internal/message"
	"n.eko.moe/neko/internal/session"
)

func (h *MessageHandler) identityDetails(id string, session *session.Session, payload *message.IdentityDetails) error {
	if _, err := h.sessions.SetName(id, payload.Username); err != nil {
		return err
	}
	return nil
}
