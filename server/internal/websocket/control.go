package websocket

import (
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
)

func (h *MessageHandler) controlRelease(id string, session types.Session) error {
	// check if session is host
	if !h.sessions.IsHost(id) {
		h.logger.Debug().Str("id", id).Msg("is not the host")
		return nil
	}

	// release host
	h.logger.Debug().Str("id", id).Msgf("host called %s", event.CONTROL_RELEASE)
	h.sessions.ClearHost()

	// tell everyone
	if err := h.sessions.Broadcast(
		message.Control{
			Event: event.CONTROL_RELEASE,
			ID:    id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_RELEASE)
		return err
	}

	return nil
}

func (h *MessageHandler) controlRequest(id string, session types.Session) error {
	// check for host
	if !h.sessions.HasHost() {
		// check if control is locked or user is admin
		_, ok := h.locked["control"]
		if ok && !session.Admin() {
			h.logger.Debug().Msg("control is locked")
			return nil
		}

		// set host
		err := h.sessions.SetHost(id)
		if err != nil {
			return err
		}

		// let everyone know
		if err := h.sessions.Broadcast(
			message.Control{
				Event: event.CONTROL_LOCKED,
				ID:    id,
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_LOCKED)
			return err
		}

		return nil
	}

	// get host
	host, ok := h.sessions.GetHost()
	if ok {

		// tell session there is a host
		if err := session.Send(message.Control{
			Event: event.CONTROL_REQUEST,
			ID:    host.ID(),
		}); err != nil {
			h.logger.Warn().Err(err).Str("id", id).Msgf("sending event %s has failed", event.CONTROL_REQUEST)
			return err
		}

		// tell host session wants to be host
		if err := host.Send(message.Control{
			Event: event.CONTROL_REQUESTING,
			ID:    id,
		}); err != nil {
			h.logger.Warn().Err(err).Str("id", host.ID()).Msgf("sending event %s has failed", event.CONTROL_REQUESTING)
			return err
		}
	}

	return nil
}

func (h *MessageHandler) controlGive(id string, session types.Session, payload *message.Control) error {
	// check if session is host
	if !h.sessions.IsHost(id) {
		h.logger.Debug().Str("id", id).Msg("is not the host")
		return nil
	}

	if !h.sessions.Has(payload.ID) {
		h.logger.Debug().Str("id", payload.ID).Msg("user does not exist")
		return nil
	}

	// check if control is locked or giver is admin
	_, ok := h.locked["control"]
	if ok && !session.Admin() {
		h.logger.Debug().Msg("control is locked")
		return nil
	}

	// set host
	err := h.sessions.SetHost(payload.ID)
	if err != nil {
		return err
	}

	// let everyone know
	if err := h.sessions.Broadcast(
		message.ControlTarget{
			Event:  event.CONTROL_GIVE,
			ID:     id,
			Target: payload.ID,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("broadcasting event %s has failed", event.CONTROL_LOCKED)
		return err
	}

	return nil
}

func (h *MessageHandler) controlClipboard(id string, session types.Session, payload *message.Clipboard) error {
	// check if session is host
	if !h.sessions.IsHost(id) {
		h.logger.Debug().Str("id", id).Msg("is not the host")
		return nil
	}

	h.remote.WriteClipboard(payload.Text)
	return nil
}

func (h *MessageHandler) controlKeyboard(id string, session types.Session, payload *message.Keyboard) error {
	// check if session is host
	if !h.sessions.IsHost(id) {
		h.logger.Debug().Str("id", id).Msg("is not the host")
		return nil
	}

	// change layout
	if payload.Layout != nil {
		h.remote.SetKeyboardLayout(*payload.Layout)
	}

	// set num lock
	var NumLock = 0
	if payload.NumLock == nil {
		NumLock = -1
	} else if *payload.NumLock {
		NumLock = 1
	}

	// set caps lock
	var CapsLock = 0
	if payload.CapsLock == nil {
		CapsLock = -1
	} else if *payload.CapsLock {
		CapsLock = 1
	}

	// set scroll lock
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

	h.remote.SetKeyboardModifiers(NumLock, CapsLock, ScrollLock)
	return nil
}
