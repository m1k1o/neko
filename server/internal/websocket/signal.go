package websocket

import (
	"n.eko.moe/neko/internal/types"
	"n.eko.moe/neko/internal/types/event"
	"n.eko.moe/neko/internal/types/message"
)

func (h *MessageHandler) signalProvide(id string, session types.Session) error {
	sdp, err := h.webrtc.CreatePeer(id, session)
	if err != nil {
		return err
	}

	if err := session.Send(message.SignalProvide{
		Event: event.SIGNAL_PROVIDE,
		ID:    id,
		SDP:   sdp,
	}); err != nil {
		return err
	}

	return nil
}

func (h *MessageHandler) signalAnswer(id string, session types.Session, payload *message.SignalAnswer) error {
	if err := session.SetName(payload.DisplayName); err != nil {
		return err
	}

	if err := session.SignalAnswer(payload.SDP); err != nil {
		return err
	}

	return nil
}
