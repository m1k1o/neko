package handler

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
		if h.state.IsLocked("control") && !session.Admin() {
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
	if h.state.IsLocked("control") && !session.Admin() {
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
	// check if session can access clipboard
	if (!h.webrtc.ImplicitControl() && !h.sessions.IsHost(id)) || (h.webrtc.ImplicitControl() && !h.sessions.CanControl(id)) {
		h.logger.Debug().Str("id", id).Msg("cannot access clipboard")
		return nil
	}

	h.desktop.WriteClipboard(payload.Text)
	return nil
}

func (h *MessageHandler) controlKeyboard(id string, session types.Session, payload *message.Keyboard) error {
	// check if session can control keyboard
	if (!h.webrtc.ImplicitControl() && !h.sessions.IsHost(id)) || (h.webrtc.ImplicitControl() && !h.sessions.CanControl(id)) {
		h.logger.Debug().Str("id", id).Msg("cannot control keyboard")
		return nil
	}

	h.desktop.SetKeyboardModifiers(types.KeyboardModifiers{
		NumLock:  payload.NumLock,
		CapsLock: payload.CapsLock,
		// TODO: ScrollLock is deprecated.
	})

	// change layout
	if payload.Layout != nil {
		return h.desktop.SetKeyboardMap(types.KeyboardMap{
			Layout: *payload.Layout,
		})
	}

	return nil
}
