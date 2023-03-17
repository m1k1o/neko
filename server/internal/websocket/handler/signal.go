package handler

import (
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/event"
	"m1k1o/neko/internal/types/message"
)

func (h *MessageHandler) signalProvide(id string, session types.Session) error {
	peer, err := h.webrtc.CreatePeer(id, session)
	if err != nil {
		return err
	}

	sdp, err := peer.CreateOffer()
	if err != nil {
		return err
	}

	if err := session.Send(message.SignalProvide{
		Event: event.SIGNAL_PROVIDE,
		ID:    id,
		SDP:   sdp,
		Lite:  h.webrtc.ICELite(),
		ICE:   h.webrtc.ICEServers(),
	}); err != nil {
		return err
	}

	return nil
}

func (h *MessageHandler) signalRemoteOffer(id string, session types.Session, payload *message.SignalOffer) error {
	return session.SignalRemoteOffer(payload.SDP)
}

func (h *MessageHandler) signalRemoteAnswer(id string, session types.Session, payload *message.SignalAnswer) error {
	if err := session.SetName(payload.DisplayName); err != nil {
		return err
	}

	if err := session.SignalRemoteAnswer(payload.SDP); err != nil {
		return err
	}

	return nil
}

func (h *MessageHandler) signalRemoteCandidate(id string, session types.Session, payload *message.SignalCandidate) error {
	return session.SignalRemoteCandidate(payload.Data)
}
