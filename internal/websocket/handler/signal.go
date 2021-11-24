package handler

import (
	"errors"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) signalRequest(session types.Session, payload *message.SignalVideo) error {
	if !session.Profile().CanWatch {
		return errors.New("not allowed to watch")
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

	session.Send(
		event.SIGNAL_PROVIDE,
		message.SignalProvide{
			SDP:        offer.SDP,
			ICEServers: h.webrtc.ICEServers(),
			Video:      payload.Video,
		})

	return nil
}

func (h *MessageHandlerCtx) signalRestart(session types.Session) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return errors.New("webRTC peer does not exist")
	}

	offer, err := peer.CreateOffer(true)
	if err != nil {
		return err
	}

	session.Send(
		event.SIGNAL_RESTART,
		message.SignalDescription{
			SDP: offer.SDP,
		})

	return nil
}

func (h *MessageHandlerCtx) signalOffer(session types.Session, payload *message.SignalDescription) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return errors.New("webRTC peer does not exist")
	}

	return peer.SignalOffer(payload.SDP)
}

func (h *MessageHandlerCtx) signalAnswer(session types.Session, payload *message.SignalDescription) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return errors.New("webRTC peer does not exist")
	}

	return peer.SignalAnswer(payload.SDP)
}

func (h *MessageHandlerCtx) signalCandidate(session types.Session, payload *message.SignalCandidate) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return errors.New("webRTC peer does not exist")
	}

	return peer.SignalCandidate(payload.ICECandidateInit)
}

func (h *MessageHandlerCtx) signalVideo(session types.Session, payload *message.SignalVideo) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return errors.New("webRTC peer does not exist")
	}

	err := peer.SetVideoID(payload.Video)
	if err != nil {
		return err
	}

	session.Send(
		event.SIGNAL_VIDEO,
		message.SignalVideo{
			Video: payload.Video,
		})

	return nil
}
