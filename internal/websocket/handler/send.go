package handler

import (
	"errors"

	"gitlab.com/demodesk/neko/server/pkg/types"
	"gitlab.com/demodesk/neko/server/pkg/types/event"
	"gitlab.com/demodesk/neko/server/pkg/types/message"
)

func (h *MessageHandlerCtx) sendUnicast(session types.Session, payload *message.SendUnicast) error {
	receiver, ok := h.sessions.Get(payload.Receiver)
	if !ok {
		return errors.New("receiver session ID not found")
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
