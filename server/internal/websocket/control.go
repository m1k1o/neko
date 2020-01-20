package websocket

import (
	"n.eko.moe/neko/internal/event"
	"n.eko.moe/neko/internal/message"
	"n.eko.moe/neko/internal/session"
)

func (h *MessageHandler) controlRelease(id string, session *session.Session) error {

	// check if session is host
	if !h.sessions.IsHost(id) {
		return nil
	}

	// release host
	h.logger.Debug().Str("id", id).Msgf("host called %s", event.CONTROL_RELEASE)
	h.sessions.ClearHost()

	// tell everyone
	if err := h.sessions.Brodcast(
		message.Control{
			Event: event.CONTROL_RELEASE,
			ID:    id,
		}, nil); err != nil {
		h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.CONTROL_RELEASE)
		return err
	}

	return nil
}

func (h *MessageHandler) controlRequest(id string, session *session.Session) error {
	h.logger.Debug().Str("id", id).Msgf("user called %s", event.CONTROL_REQUEST)

	// check for host
	if !h.sessions.HasHost() {
		// set host
		h.sessions.SetHost(id)

		// let everyone know
		if err := h.sessions.Brodcast(
			message.Control{
				Event: event.CONTROL_LOCKED,
				ID:    id,
			}, nil); err != nil {
			h.logger.Warn().Err(err).Msgf("brodcasting event %s has failed", event.CONTROL_LOCKED)
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
			ID:    host.ID,
		}); err != nil {
			h.logger.Warn().Err(err).Str("id", id).Msgf("sending event %s has failed", event.CONTROL_REQUEST)
			return err
		}

		// tell host session wants to be host
		if err := host.Send(message.Control{
			Event: event.CONTROL_REQUESTING,
			ID:    id,
		}); err != nil {
			h.logger.Warn().Err(err).Str("id", host.ID).Msgf("sending event %s has failed", event.CONTROL_REQUESTING)
			return err
		}
	}

	return nil
}
