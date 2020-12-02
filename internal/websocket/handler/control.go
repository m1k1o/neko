package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) controlRelease(session types.Session) error {
	if !session.IsHost() {
		h.logger.Debug().Str("id", session.ID()).Msg("is not the host")
		return nil
	}

	h.desktop.ResetKeys()
	h.sessions.ClearHost()

	h.sessions.Broadcast(
		message.ControlHost{
			Event:   event.CONTROL_HOST,
			HasHost: false,
		}, nil)

	return nil
}

func (h *MessageHandlerCtx) controlRequest(session types.Session) error {
	if session.IsHost() {
		h.logger.Debug().Str("id", session.ID()).Msg("is already the host")
		return nil
	}

	if !h.sessions.ImplicitHosting() {
		// tell session if there is a host
		if host := h.sessions.GetHost(); host != nil {
			return session.Send(
				message.ControlHost{
					Event:   event.CONTROL_HOST,
					HasHost: true,
					HostID:  host.ID(),
				})
		}
	}

	h.sessions.SetHost(session)
	h.sessions.Broadcast(
		message.ControlHost{
			Event:   event.CONTROL_HOST,
			HasHost: true,
			HostID:  session.ID(),
		}, nil)

	return nil
}
