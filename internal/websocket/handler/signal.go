package handler

import (
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) signalRequest(session types.Session, payload *message.SignalVideo) error {
	if !session.Profile().CanWatch {
		h.logger.Debug().Str("session_id", session.ID()).Msg("not allowed to watch")
		return nil
	}

	// use default first video, if not provided
	if payload.Video == "" {
		videos := h.capture.VideoIDs()
		payload.Video = videos[0]
	}

	offer, err := h.webrtc.CreatePeer(session, payload.Video)
	if err != nil {
		return err
	}

	return session.Send(
		message.SignalProvide{
			Event:      event.SIGNAL_PROVIDE,
			SDP:        offer.SDP,
			ICEServers: h.webrtc.ICEServers(),
			Video:      payload.Video,
		})
}

func (h *MessageHandlerCtx) signalRestart(session types.Session) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		h.logger.Debug().Str("session_id", session.ID()).Msg("webRTC peer does not exist")
		return nil
	}

	offer, err := peer.CreateOffer(true)
	if err != nil {
		return err
	}

	return session.Send(
		message.SignalAnswer{
			Event: event.SIGNAL_RESTART,
			SDP:   offer.SDP,
		})
}

func (h *MessageHandlerCtx) signalAnswer(session types.Session, payload *message.SignalAnswer) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		h.logger.Debug().Str("session_id", session.ID()).Msg("webRTC peer does not exist")
		return nil
	}

	return peer.SignalAnswer(payload.SDP)
}

func (h *MessageHandlerCtx) signalCandidate(session types.Session, payload *message.SignalCandidate) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		h.logger.Debug().Str("session_id", session.ID()).Msg("webRTC peer does not exist")
		return nil
	}

	return peer.SignalCandidate(*payload.ICECandidateInit)
}

func (h *MessageHandlerCtx) signalVideo(session types.Session, payload *message.SignalVideo) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		h.logger.Debug().Str("session_id", session.ID()).Msg("webRTC peer does not exist")
		return nil
	}

	err := peer.SetVideoID(payload.Video)
	if err != nil {
		return err
	}

	return session.Send(
		message.SignalVideo{
			Event: event.SIGNAL_VIDEO,
			Video: payload.Video,
		})
}
