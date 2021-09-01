package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) controlRelease(session types.Session) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.Profile().CanHost {
		logger.Debug().Msg("is not allowed to host")
		return nil
	}

	if !session.IsHost() {
		logger.Debug().Msg("is not the host")
		return nil
	}

	h.desktop.ResetKeys()
	h.sessions.ClearHost()

	return nil
}

func (h *MessageHandlerCtx) controlRequest(session types.Session) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.Profile().CanHost {
		logger.Debug().Msg("is not allowed to host")
		return nil
	}

	if session.IsHost() {
		logger.Debug().Msg("is already the host")
		return nil
	}

	if !h.sessions.ImplicitHosting() {
		// tell session if there is a host
		if host := h.sessions.GetHost(); host != nil {
			session.Send(
				event.CONTROL_HOST,
				message.ControlHost{
					HasHost: true,
					HostID:  host.ID(),
				})

			return nil
		}
	}

	h.sessions.SetHost(session)

	return nil
}
