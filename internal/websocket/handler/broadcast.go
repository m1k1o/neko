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
	return h.boradcastStatus(session)
}

func (h *MessageHandlerCtx) boradcastDestroy(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	h.capture.StopBroadcast()
	return h.boradcastStatus(session)
}

func (h *MessageHandlerCtx) boradcastStatus(session types.Session) error {
	if !session.Admin() {
		h.logger.Debug().Msg("user not admin")
		return nil
	}

	return session.Send(
		message.BroadcastStatus{
			Event:    event.BORADCAST_STATUS,
			IsActive: h.capture.BroadcastEnabled(),
			URL:      h.capture.BroadcastUrl(),
		})
}
