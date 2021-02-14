package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) sendUnicast(session types.Session, payload *message.SendUnicast) error {
	receiver, ok := h.sessions.Get(payload.Receiver)
	if !ok {
		h.logger.Debug().Str("id", session.ID()).Msg("receiver ID not found")
		return nil
	}

	return receiver.Send(message.SendUnicast{
		Event:    event.SEND_UNICAST,
		Sender:   session.ID(),
		Receiver: receiver.ID(),
		Subject:  payload.Subject,
		Body:     payload.Body,
	})
}

func (h *MessageHandlerCtx) sendBroadcast(session types.Session, payload *message.SendBroadcast) error {
	h.sessions.Broadcast(message.SendBroadcast{
		Event:   event.SEND_BROADCAST,
		Sender:  session.ID(),
		Subject: payload.Subject,
		Body:    payload.Body,
	}, []string{session.ID()})
	return nil
}
