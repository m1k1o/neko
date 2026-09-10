package handler

import (
	signalingapp "github.com/m1k1o/neko/server/internal/application/signaling"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

func (h *MessageHandlerCtx) signalRequest(session types.Session, payload *message.SignalRequest) error {
	provide, err := h.signaling.Request(session, signalingapp.Request{
		Video:       payload.Video,
		Audio:       payload.Audio,
		VideoCodecs: payload.VideoCodecs,
	})
	if err != nil {
		return err
	}

	h.logger.Info().
		Str("session_id", session.ID()).
		Str("video_codec", provide.VideoCodec).
		Strs("browser_video_codecs", payload.VideoCodecs).
		Msg("selected video codec for browser capabilities")

	session.Send(event.SIGNAL_PROVIDE, message.SignalProvide{
		SDP:        provide.SDP,
		ICEServers: provide.ICEServers,
		Video:      provide.Video,
		Audio:      provide.Audio,
	})
	return nil
}

func (h *MessageHandlerCtx) signalRestart(session types.Session) error {
	sdp, err := h.signaling.Restart(session)
	if err != nil {
		return err
	}
	session.Send(event.SIGNAL_RESTART, message.SignalDescription{SDP: sdp})
	return nil
}

func (h *MessageHandlerCtx) signalOffer(session types.Session, payload *message.SignalDescription) error {
	sdp, err := h.signaling.Offer(session, payload.SDP)
	if err != nil {
		return err
	}
	session.Send(event.SIGNAL_ANSWER, message.SignalDescription{SDP: sdp})
	return nil
}

func (h *MessageHandlerCtx) signalAnswer(session types.Session, payload *message.SignalDescription) error {
	return h.signaling.Answer(session, payload.SDP)
}

func (h *MessageHandlerCtx) signalCandidate(session types.Session, payload *message.SignalCandidate) error {
	return h.signaling.Candidate(session, payload.ICECandidateInit)
}

func (h *MessageHandlerCtx) signalVideo(session types.Session, payload *message.SignalVideo) error {
	return h.signaling.Video(session, payload.PeerVideoRequest)
}

func (h *MessageHandlerCtx) signalAudio(session types.Session, payload *message.SignalAudio) error {
	return h.signaling.Audio(session, payload.PeerAudioRequest)
}
