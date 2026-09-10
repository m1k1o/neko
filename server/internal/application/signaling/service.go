package signaling

import (
	"errors"
	"fmt"

	"github.com/pion/webrtc/v4"

	"github.com/m1k1o/neko/server/pkg/types"
)

type Request struct {
	Video       types.PeerVideoRequest
	Audio       types.PeerAudioRequest
	VideoCodecs []string
}

type Provide struct {
	SDP        string
	ICEServers []types.ICEServer
	Video      types.PeerVideo
	Audio      types.PeerAudio
	VideoCodec string
}

type Service struct {
	capture types.CaptureManager
	webrtc  types.WebRTCManager
}

func NewService(capture types.CaptureManager, webrtc types.WebRTCManager) *Service {
	return &Service{capture: capture, webrtc: webrtc}
}

func (s *Service) Request(session types.Session, request Request) (Provide, error) {
	if !session.Profile().CanWatch {
		return Provide{}, errors.New("not allowed to watch")
	}

	videoCodec, ok := s.capture.SelectVideoCodec(request.VideoCodecs)
	if !ok {
		return Provide{}, fmt.Errorf("none of the browser video codecs are available on the server")
	}

	offer, peer, err := s.webrtc.CreatePeer(session, videoCodec)
	if err != nil {
		return Provide{}, err
	}
	if session.PrivateModeEnabled() {
		peer.SetPaused(true)
	}

	video := request.Video
	if video.Selector == nil {
		videoManager, ok := s.capture.VideoForCodec(videoCodec)
		if !ok {
			return Provide{}, fmt.Errorf("video codec %q is not available on the server", videoCodec.Name)
		}
		videos := videoManager.IDs()
		if len(videos) == 0 {
			return Provide{}, errors.New("no video streams are configured")
		}
		video.Selector = &types.StreamSelector{ID: videos[0], Type: types.StreamSelectorTypeExact}
	}

	if err := peer.SetVideo(video); err != nil {
		return Provide{}, err
	}

	audio := request.Audio
	if audio.Disabled == nil {
		disabled := false
		audio.Disabled = &disabled
	}
	if err := peer.SetAudio(audio); err != nil {
		return Provide{}, err
	}

	return Provide{
		SDP:        offer.SDP,
		ICEServers: s.webrtc.ICEServers(),
		Video:      peer.Video(),
		Audio:      peer.Audio(),
		VideoCodec: videoCodec.Name,
	}, nil
}

func (s *Service) Restart(session types.Session) (string, error) {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return "", errors.New("webRTC peer does not exist")
	}
	offer, err := peer.CreateOffer(true)
	if err != nil {
		return "", err
	}
	return offer.SDP, nil
}

func (s *Service) Offer(session types.Session, sdp string) (string, error) {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return "", errors.New("webRTC peer does not exist")
	}
	if err := peer.SetRemoteDescription(webrtc.SessionDescription{SDP: sdp, Type: webrtc.SDPTypeOffer}); err != nil {
		return "", err
	}
	answer, err := peer.CreateAnswer()
	if err != nil {
		return "", err
	}
	return answer.SDP, nil
}

func (s *Service) Answer(session types.Session, sdp string) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return errors.New("webRTC peer does not exist")
	}
	return peer.SetRemoteDescription(webrtc.SessionDescription{SDP: sdp, Type: webrtc.SDPTypeAnswer})
}

func (s *Service) Candidate(session types.Session, candidate webrtc.ICECandidateInit) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return errors.New("webRTC peer does not exist")
	}
	return peer.SetCandidate(candidate)
}

func (s *Service) Video(session types.Session, request types.PeerVideoRequest) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return errors.New("webRTC peer does not exist")
	}
	return peer.SetVideo(request)
}

func (s *Service) Audio(session types.Session, request types.PeerAudioRequest) error {
	peer := session.GetWebRTCPeer()
	if peer == nil {
		return errors.New("webRTC peer does not exist")
	}
	return peer.SetAudio(request)
}
