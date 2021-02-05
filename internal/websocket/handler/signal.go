package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) signalRequest(session types.Session) error {
	if !session.CanWatch() {
		return nil
	}

	offer, err := h.webrtc.CreatePeer(session)
	if err != nil {
		return err
	}

	return session.Send(
		message.SignalProvide{
			Event: event.SIGNAL_PROVIDE,
			SDP:   offer.SDP,
			Lite:  h.webrtc.ICELite(),
			ICE:   h.webrtc.ICEServers(),
		})
}

func (h *MessageHandlerCtx) signalAnswer(session types.Session, payload *message.SignalAnswer) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		h.logger.Debug().Msg("webRTC peer does not exist")
		return nil
	}

	return peer.SignalAnswer(payload.SDP)
}

func (h *MessageHandlerCtx) signalCandidate(session types.Session, payload *message.SignalCandidate) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		h.logger.Debug().Msg("webRTC peer does not exist")
		return nil
	}

	return peer.SignalCandidate(*payload.ICECandidateInit)
}

func (h *MessageHandlerCtx) signalVideo(session types.Session, payload *message.SignalVideo) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		h.logger.Debug().Msg("webRTC peer does not exist")
		return nil
	}

	return peer.SetVideoID(payload.VideoID)
}
