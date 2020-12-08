package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) SessionCreated(session types.Session) error {
	h.sessions.Broadcast(
		message.MemberData{
			Event:   event.MEMBER_CREATED,
			ID:      session.ID(),
			Profile: session.GetProfile(),
			State:   session.GetState(),
		}, nil)

	return nil
}

func (h *MessageHandlerCtx) SessionDeleted(session types.Session) error {
	h.sessions.Broadcast(
		message.MemberID{
			Event: event.MEMBER_DELETED,
			ID:    session.ID(),
		}, nil);

	return nil
}

func (h *MessageHandlerCtx) SessionConnected(session types.Session) error {
	// start streaming, when first member connects
	if !h.capture.Streaming() {
		h.capture.StartStream()
	}

	if err := h.systemInit(session); err != nil {
		return err
	}

	if session.IsAdmin() {
		if err := h.systemAdmin(session); err != nil {
			return err
		}
	}

	return h.SessionStateChanged(session)
}

func (h *MessageHandlerCtx) SessionDisconnected(session types.Session) error {
	// Stop streaming, if last member disonnects
	if h.capture.Streaming() && !h.sessions.HasConnectedMembers() {
		h.capture.StopStream()
	}

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

	return h.SessionStateChanged(session)
}

func (h *MessageHandlerCtx) SessionProfileChanged(session types.Session) error {
	profile := session.GetProfile()

	h.sessions.Broadcast(
		message.MemberProfile{
			Event:         event.MEMBER_PROFILE,
			ID:            session.ID(),
			MemberProfile: &profile,
		}, nil)

	return nil
}

func (h *MessageHandlerCtx) SessionStateChanged(session types.Session) error {
	state := session.GetState()

	h.sessions.Broadcast(
		message.MemberState{
			Event:       event.MEMBER_STATE,
			ID:          session.ID(),
			MemberState: &state,
		}, nil)

	return nil
}
