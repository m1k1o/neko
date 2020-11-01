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

	h.logger.Debug().Str("id", session.ID()).Msgf("host called %s", event.CONTROL_RELEASE)
	h.sessions.ClearHost()

	if err := h.sessions.Broadcast(
		message.Control{
			Event: event.CONTROL_RELEASE,
			ID:    session.ID(),
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_RELEASE)
		return err
	}

	return nil
}

func (h *MessageHandlerCtx) controlRequest(session types.Session) error {
	host := h.sessions.GetHost()

	if host == nil {
		// set host
		h.sessions.SetHost(session)

		// let everyone know
		if err := h.sessions.Broadcast(
			message.Control{
				Event: event.CONTROL_LOCKED,
				ID:    session.ID(),
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_LOCKED)
			return err
		}
	} else {
		// tell session there is a host
		if err := session.Send(message.Control{
			Event: event.CONTROL_REQUEST,
			ID:    host.ID(),
		}); err != nil {
			h.logger.Warn().Err(err).Str("id", session.ID()).Msgf("sending event %s has failed", event.CONTROL_REQUEST)
			return err
		}

		// tell host session wants to be host
		if err := host.Send(message.Control{
			Event: event.CONTROL_REQUESTING,
			ID:    session.ID(),
		}); err != nil {
			h.logger.Warn().Err(err).Str("id", host.ID()).Msgf("sending event %s has failed", event.CONTROL_REQUESTING)
			return err
		}
	}

	return nil
}

func (h *MessageHandlerCtx) controlGive(session types.Session, payload *message.Control) error {
	if !session.IsHost() {
		h.logger.Debug().Str("id", session.ID()).Msg("is not the host")
		return nil
	}

	target, ok := h.sessions.Get(payload.ID)
	if !ok {
		h.logger.Debug().Str("id", payload.ID).Msg("can't find target session")
		return nil
	}

	h.sessions.SetHost(target)

	if err := h.sessions.Broadcast(
		message.ControlTarget{
			Event:  event.CONTROL_GIVE,
			ID:     session.ID(),
			Target: target.ID(),
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_LOCKED)
		return err
	}

	return nil
}

func (h *MessageHandlerCtx) controlClipboard(session types.Session, payload *message.Clipboard) error {
	if !session.IsHost() {
		h.logger.Debug().Str("id", session.ID()).Msg("is not the host")
		return nil
	}

	h.desktop.WriteClipboard(payload.Text)
	return nil
}

func (h *MessageHandlerCtx) controlKeyboard(session types.Session, payload *message.Keyboard) error {
	if !session.IsHost() {
		h.logger.Debug().Str("id", session.ID()).Msg("is not the host")
		return nil
	}

	// change layout
	if payload.Layout != nil {
		h.desktop.SetKeyboardLayout(*payload.Layout)
	}

	var NumLock = 0
	if payload.NumLock == nil {
		NumLock = -1
	} else if *payload.NumLock {
		NumLock = 1
	}

	var CapsLock = 0
	if payload.CapsLock == nil {
		CapsLock = -1
	} else if *payload.CapsLock {
		CapsLock = 1
	}

	var ScrollLock = 0
	if payload.ScrollLock == nil {
		ScrollLock = -1
	} else if *payload.ScrollLock {
		ScrollLock = 1
	}

	h.logger.Debug().
		Int("NumLock", NumLock).
		Int("CapsLock", CapsLock).
		Int("ScrollLock", ScrollLock).
		Msg("setting keyboard modifiers")

	h.desktop.SetKeyboardModifiers(NumLock, CapsLock, ScrollLock)
	return nil
}
