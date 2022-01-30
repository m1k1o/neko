package handler

import (
	"errors"

	"demodesk/neko/internal/desktop/xorg"
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
	if !session.Profile().CanHost {
		return errors.New("is not allowed to host")
	}

	if session.IsHost() {
		return errors.New("is already the host")
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

func (h *MessageHandlerCtx) controlKeyPress(session types.Session, payload *message.ControlKey) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.IsHost() {
		logger.Debug().Msg("is not the host")
		return nil
	}

	return h.desktop.KeyPress(payload.Keysym)
}

func (h *MessageHandlerCtx) controlKeyDown(session types.Session, payload *message.ControlKey) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.IsHost() {
		logger.Debug().Msg("is not the host")
		return nil
	}

	return h.desktop.KeyDown(payload.Keysym)
}

func (h *MessageHandlerCtx) controlKeyUp(session types.Session, payload *message.ControlKey) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.IsHost() {
		logger.Debug().Msg("is not the host")
		return nil
	}

	return h.desktop.KeyUp(payload.Keysym)
}

func (h *MessageHandlerCtx) controlCopy(session types.Session) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.IsHost() {
		logger.Debug().Msg("is not the host")
		return nil
	}

	return h.desktop.KeyPress(xorg.XK_Control_L, xorg.XK_c)
}

func (h *MessageHandlerCtx) controlPaste(session types.Session) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.IsHost() {
		logger.Debug().Msg("is not the host")
		return nil
	}

	return h.desktop.KeyPress(xorg.XK_Control_L, xorg.XK_v)
}

func (h *MessageHandlerCtx) controlSelectAll(session types.Session) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	if !session.IsHost() {
		logger.Debug().Msg("is not the host")
		return nil
	}

	return h.desktop.KeyPress(xorg.XK_Control_L, xorg.XK_a)
}
