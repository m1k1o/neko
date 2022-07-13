package handler

import (
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
)

func (h *MessageHandlerCtx) SessionCreated(session types.Session) error {
	h.sessions.Broadcast(
		event.SESSION_CREATED,
		message.SessionData{
			ID:      session.ID(),
			Profile: session.Profile(),
			State:   session.State(),
		}, nil)

	return nil
}

func (h *MessageHandlerCtx) SessionDeleted(session types.Session) error {
	h.sessions.Broadcast(
		event.SESSION_DELETED,
		message.SessionID{
			ID: session.ID(),
		}, nil)

	return nil
}

func (h *MessageHandlerCtx) SessionConnected(session types.Session) error {
	if err := h.systemInit(session); err != nil {
		return err
	}

	if session.Profile().IsAdmin {
		if err := h.systemAdmin(session); err != nil {
			return err
		}
	}

	return h.SessionStateChanged(session)
}

func (h *MessageHandlerCtx) SessionDisconnected(session types.Session) error {
	// clear host if exists
	if session.IsHost() {
		h.desktop.ResetKeys()
		h.sessions.ClearHost()
	}

	return h.SessionStateChanged(session)
}

func (h *MessageHandlerCtx) SessionProfileChanged(session types.Session) error {
	h.sessions.Broadcast(
		event.SESSION_PROFILE,
		message.MemberProfile{
			ID:            session.ID(),
			MemberProfile: session.Profile(),
		}, nil)

	return nil
}

func (h *MessageHandlerCtx) SessionStateChanged(session types.Session) error {
	h.sessions.Broadcast(
		event.SESSION_STATE,
		message.SessionState{
			ID:           session.ID(),
			SessionState: session.State(),
		}, nil)

	return nil
}
