package webrtc

import (
	"fmt"

	"github.com/pion/webrtc/v3"
)

type WebRTCPeerCtx struct {
	api            *webrtc.API
	engine         *webrtc.MediaEngine
	settings       *webrtc.SettingEngine
	connection     *webrtc.PeerConnection
	configuration  *webrtc.Configuration
	videoTracks    map[string]*webrtc.TrackLocalStaticSample
	videoSender    *webrtc.RTPSender
}

func (webrtc_peer *WebRTCPeerCtx) SignalAnswer(sdp string) error {
	return webrtc_peer.connection.SetRemoteDescription(webrtc.SessionDescription{
		SDP: sdp,
		Type: webrtc.SDPTypeAnswer,
	})
}

func (webrtc_peer *WebRTCPeerCtx) SignalCandidate(candidate webrtc.ICECandidateInit) error {
	return webrtc_peer.connection.AddICECandidate(candidate)
}

func (webrtc_peer *WebRTCPeerCtx) SetVideoID(videoID string) error {
	track, ok := webrtc_peer.videoTracks[videoID]
	if !ok {
		return fmt.Errorf("videoID not found in available tracks")
	}

	return webrtc_peer.videoSender.ReplaceTrack(track)
}

func (webrtc_peer *WebRTCPeerCtx) Destroy() error {
	if webrtc_peer.connection == nil || webrtc_peer.connection.ConnectionState() != webrtc.PeerConnectionStateConnected {
		return nil
	}

	return webrtc_peer.connection.Close()
}
