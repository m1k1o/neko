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

	h.sessions.ClearHost()
	h.sessions.Broadcast(
		message.ControlHost{
			Event:   event.CONTROL_HOST,
			HasHost: false,
		}, nil)

	return nil
}

func (h *MessageHandlerCtx) controlRequest(session types.Session) error {
	// TODO: Allow implicit requests.
	host := h.sessions.GetHost()
	if host != nil {
		// tell session there is a host
		return session.Send(
			message.ControlHost{
				Event:   event.CONTROL_HOST,
				HasHost: true,
				HostID:  host.ID(),
			})
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
