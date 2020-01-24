package websocket

import (
	"n.eko.moe/neko/internal/types"
	"n.eko.moe/neko/internal/types/event"
	"n.eko.moe/neko/internal/types/message"
)

func (h *MessageHandler) identityDetails(id string, session types.Session, payload *message.IdentityDetails) error {
	if err := session.SetName(payload.Username); err != nil {
		return err
	}
	return nil
}

func (h *MessageHandler) createPeer(id string, session types.Session, payload *message.Signal) error {
	sdp, peer, err := h.webrtc.CreatePeer(id, payload.SDP)
	if err != nil {
		return err
	}

	if err := session.SetPeer(peer); err != nil {
		return err
	}

	if err := session.Send(message.Signal{
		Event: event.SIGNAL_ANSWER,
		SDP:   sdp,
	}); err != nil {
		return err
	}

	return nil
}
