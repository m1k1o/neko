package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) SessionCreated(session types.Session) error {
	h.sessions.Broadcast(
		message.SessionData{
			Event:   event.SESSION_CREATED,
			ID:      session.ID(),
			Profile: session.Profile(),
			State:   session.State(),
		}, nil)

	return nil
}

func (h *MessageHandlerCtx) SessionDeleted(session types.Session) error {
	h.sessions.Broadcast(
		message.SessionID{
			Event: event.SESSION_DELETED,
			ID:    session.ID(),
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
	profile := session.Profile()

	h.sessions.Broadcast(
		message.MemberProfile{
			Event:         event.SESSION_PROFILE,
			ID:            session.ID(),
			MemberProfile: &profile,
		}, nil)

	return nil
}

func (h *MessageHandlerCtx) SessionStateChanged(session types.Session) error {
	state := session.State()

	h.sessions.Broadcast(
		message.SessionState{
			Event:        event.SESSION_STATE,
			ID:           session.ID(),
			SessionState: &state,
		}, nil)

	return nil
}
