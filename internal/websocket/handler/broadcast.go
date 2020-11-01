package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) boradcastCreate(session types.Session, payload *message.BroadcastCreate) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	h.capture.StartBroadcast(payload.URL)

	if err := h.boradcastStatus(session); err != nil {
		return err
	}

	return nil
}

func (h *MessageHandlerCtx) boradcastDestroy(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	h.capture.StopBroadcast()

	if err := h.boradcastStatus(session); err != nil {
		return err
	}

	return nil
}

func (h *MessageHandlerCtx) boradcastStatus(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	if err := session.Send(
		message.BroadcastStatus{
			Event:    event.BORADCAST_STATUS,
			IsActive: h.capture.BroadcastEnabled(),
			URL:      h.capture.BroadcastUrl(),
		}); err != nil {
		h.logger.Warn().Err(err).Msgf("sending event %s has failed", event.BORADCAST_STATUS)
		return err
	}

	return nil
}
