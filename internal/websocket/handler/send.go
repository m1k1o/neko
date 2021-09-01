package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) sendUnicast(session types.Session, payload *message.SendUnicast) error {
	logger := h.logger.With().Str("session_id", session.ID()).Logger()

	receiver, ok := h.sessions.Get(payload.Receiver)
	if !ok {
		logger.Debug().Msg("receiver session ID not found")
		return nil
	}

	receiver.Send(
		event.SEND_UNICAST,
		message.SendUnicast{
			Sender:   session.ID(),
			Receiver: receiver.ID(),
			Subject:  payload.Subject,
			Body:     payload.Body,
		})

	return nil
}

func (h *MessageHandlerCtx) sendBroadcast(session types.Session, payload *message.SendBroadcast) error {
	h.sessions.Broadcast(
		event.SEND_BROADCAST,
		message.SendBroadcast{
			Sender:  session.ID(),
			Subject: payload.Subject,
			Body:    payload.Body,
		}, []string{session.ID()})

	return nil
}
