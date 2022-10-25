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
		videos := h.capture.Video().IDs()
		payload.Video = videos[0]
	}

	var err error
	if payload.Bitrate == 0 {
		// get bitrate from video id
		payload.Bitrate, err = h.capture.GetBitrateFromVideoID(payload.Video)
		if err != nil {
			return err
		}
	}

	offer, err := h.webrtc.CreatePeer(session, payload.Bitrate)
	if err != nil {
		return err
	}

	if webrtcPeer := session.GetWebRTCPeer(); webrtcPeer != nil {
		// set webrtc as paused if session has private mode enabled
		if session.PrivateModeEnabled() {
			webrtcPeer.SetPaused(true)
		}

		payload.Video = webrtcPeer.GetVideoId()
	}

	session.Send(
		event.SIGNAL_PROVIDE,
		message.SignalProvide{
			SDP:        offer.SDP,
			ICEServers: h.webrtc.ICEServers(),
			Video:      payload.Video, // TODO: Refactor.
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

	var err error
	if payload.Bitrate == 0 {
		// get bitrate from video id
		payload.Bitrate, err = h.capture.GetBitrateFromVideoID(payload.Video)
		if err != nil {
			return err
		}
	}

	if err = peer.SetVideoBitrate(payload.Bitrate); err != nil {
		return err
	}

	session.Send(
		event.SIGNAL_VIDEO,
		message.SignalVideo{
			Video:   peer.GetVideoId(), // TODO: Refactor.
			Bitrate: payload.Bitrate,
		})

	return nil
}
