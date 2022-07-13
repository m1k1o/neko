package handler

import (
	"errors"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
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

	// set webrtc as paused if session has private mode enabled
	if webrtcPeer := session.GetWebRTCPeer(); webrtcPeer != nil && session.PrivateModeEnabled() {
		webrtcPeer.SetPaused(true)
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

	// TODO: Use offer event intead.
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

	err := peer.SetOffer(payload.SDP)
	if err != nil {
		return err
	}

	answer, err := peer.CreateAnswer()
	if err != nil {
		return err
	}

	session.Send(
		event.SIGNAL_ANSWER,
		message.SignalDescription{
			SDP: answer.SDP,
		})

	return nil
}

func (h *MessageHandlerCtx) signalAnswer(session types.Session, payload *message.SignalDescription) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return errors.New("webRTC peer does not exist")
	}

	return peer.SetAnswer(payload.SDP)
}

func (h *MessageHandlerCtx) signalCandidate(session types.Session, payload *message.SignalCandidate) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return errors.New("webRTC peer does not exist")
	}

	return peer.SetCandidate(payload.ICECandidateInit)
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
