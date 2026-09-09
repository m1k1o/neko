package handler

import (
	"errors"
	"fmt"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/types/message"

	"github.com/pion/webrtc/v4"
)

func (h *MessageHandlerCtx) signalRequest(session types.Session, payload *message.SignalRequest) error {
	if !session.Profile().CanWatch {
		return errors.New("not allowed to watch")
	}

	videoCodec, ok := h.capture.SelectVideoCodec(payload.VideoCodecs)
	if !ok {
		return fmt.Errorf("none of the browser video codecs are available on the server")
	}
	h.logger.Info().
		Str("session_id", session.ID()).
		Str("video_codec", videoCodec.Name).
		Strs("browser_video_codecs", payload.VideoCodecs).
		Msg("selected video codec for browser capabilities")

	offer, peer, err := h.webrtc.CreatePeer(session, videoCodec)
	if err != nil {
		return err
	}

	// set webrtc as paused if session has private mode enabled
	if session.PrivateModeEnabled() {
		peer.SetPaused(true)
	}

	video := payload.Video

	// use default first video, if not provided
	if video.Selector == nil {
		videoManager, ok := h.capture.VideoForCodec(videoCodec)
		if !ok {
			return fmt.Errorf("video codec %q is not available on the server", videoCodec.Name)
		}
		videos := videoManager.IDs()
		if len(videos) == 0 {
			return errors.New("no video streams are configured")
		}
		video.Selector = &types.StreamSelector{
			ID:   videos[0],
			Type: types.StreamSelectorTypeExact,
		}
	}

	// set video stream
	err = peer.SetVideo(video)
	if err != nil {
		return err
	}

	audio := payload.Audio

	// enable by default if not requested otherwise
	if audio.Disabled == nil {
		disabled := false
		audio.Disabled = &disabled
	}

	// set audio stream
	err = peer.SetAudio(audio)
	if err != nil {
		return err
	}

	session.Send(
		event.SIGNAL_PROVIDE,
		message.SignalProvide{
			SDP:        offer.SDP,
			ICEServers: h.webrtc.ICEServers(),

			Video: peer.Video(),
			Audio: peer.Audio(),
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

	// TODO: Use offer event instead.
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

	err := peer.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  payload.SDP,
		Type: webrtc.SDPTypeOffer,
	})
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

	return peer.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  payload.SDP,
		Type: webrtc.SDPTypeAnswer,
	})
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

	return peer.SetVideo(payload.PeerVideoRequest)
}

func (h *MessageHandlerCtx) signalAudio(session types.Session, payload *message.SignalAudio) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return errors.New("webRTC peer does not exist")
	}

	return peer.SetAudio(payload.PeerAudioRequest)
}
