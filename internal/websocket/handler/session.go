package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) SessionConnected(session types.Session) error {
	if err := h.systemInit(session); err != nil {
		return err
	}

	if session.IsAdmin() {
		if err := h.systemAdmin(session); err != nil {
			return err
		}
	}

	// let everyone know there is a new session
	h.sessions.Broadcast(
		message.MemberData{
			Event:   event.MEMBER_CONNECTED,
			ID:      session.ID(),
			Name:    session.Name(),
			IsAdmin: session.IsAdmin(),
		}, nil)

	return nil
}

func (h *MessageHandlerCtx) SessionDisconnected(session types.Session) error {
	// clear host if exists
	if session.IsHost() {
		h.desktop.ResetKeys()
		h.sessions.ClearHost()

		h.sessions.Broadcast(
			message.ControlHost{
				Event:   event.CONTROL_HOST,
				HasHost: false,
			}, nil)
	}

	// let everyone know session disconnected
	h.sessions.Broadcast(
		message.MemberID{
			Event: event.MEMBER_DISCONNECTED,
			ID:    session.ID(),
		}, nil);

	return nil
}
