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

	h.sessions.Broadcast(
		message.Control{
			Event: event.CONTROL_RELEASE,
			ID:    session.ID(),
		}, nil)

	return nil
}

func (h *MessageHandlerCtx) controlRequest(session types.Session) error {
	host := h.sessions.GetHost()

	if host == nil {
		// set host
		h.sessions.SetHost(session)

		// let everyone know
		h.sessions.Broadcast(
			message.Control{
				Event: event.CONTROL_LOCKED,
				ID:    session.ID(),
			}, nil)
	} else {
		// tell session there is a host
		if err := session.Send(
			message.Control{
				Event: event.CONTROL_REQUEST,
				ID:    host.ID(),
			}); err != nil {
			return err
		}

		// tell host session wants to be host
		return host.Send(
			message.Control{
				Event: event.CONTROL_REQUESTING,
				ID:    session.ID(),
			})
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

	h.sessions.Broadcast(
		message.ControlTarget{
			Event:  event.CONTROL_GIVE,
			ID:     session.ID(),
			Target: target.ID(),
		}, nil)

	return nil
}
