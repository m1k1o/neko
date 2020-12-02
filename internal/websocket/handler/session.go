package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) SessionCreated(session types.Session) error {
	// TODO: Join structs?
	h.sessions.Broadcast(
		message.MemberData{
			Event:       event.MEMBER_CREATED,
			ID:          session.ID(),
			Profile:     message.MemberProfile{
				Name:               session.Name(),
				IsAdmin:            session.IsAdmin(),
				CanLogin:           session.CanLogin(),
				CanConnect:         session.CanConnect(),
				CanWatch:           session.CanWatch(),
				CanHost:            session.CanHost(),
				CanAccessClipboard: session.CanAccessClipboard(),
			},
			IsConnected: session.IsConnected(),
			IsReceiving: session.IsReceiving(),
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

	h.sessions.Broadcast(
		message.MemberID{
			Event: event.MEMBER_CONNECTED,
			ID:    session.ID(),
		}, nil);

	return nil
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

	h.sessions.Broadcast(
		message.MemberID{
			Event: event.MEMBER_DISCONNECTED,
			ID:    session.ID(),
		}, nil);

	return nil
}

func (h *MessageHandlerCtx) SessionReceivingStarted(session types.Session) error {
	h.sessions.Broadcast(
		message.MemberID{
			Event: event.MEMBER_RECEIVING_STARTED,
			ID:    session.ID(),
		}, nil);

	return nil
}

func (h *MessageHandlerCtx) SessionReceivingStopped(session types.Session) error {
	h.sessions.Broadcast(
		message.MemberID{
			Event: event.MEMBER_RECEIVING_STOPPED,
			ID:    session.ID(),
		}, nil);

	return nil
}

func (h *MessageHandlerCtx) SessionProfileUpdated(session types.Session) error {
	// TODO: Join structs?
	h.sessions.Broadcast(
		message.MemberProfile{
			Event:              event.MEMBER_PROFILE_UPDATED,
			ID:                 session.ID(),
			Name:               session.Name(),
			IsAdmin:            session.IsAdmin(),
			CanLogin:           session.CanLogin(),
			CanConnect:         session.CanConnect(),
			CanWatch:           session.CanWatch(),
			CanHost:            session.CanHost(),
			CanAccessClipboard: session.CanAccessClipboard(),
		}, nil)

	return nil
}
