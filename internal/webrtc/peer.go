package webrtc

import "github.com/pion/webrtc/v3"

type WebRTCPeerCtx struct {
	api               *webrtc.API
	engine            *webrtc.MediaEngine
	settings          *webrtc.SettingEngine
	connection        *webrtc.PeerConnection
	configuration     *webrtc.Configuration
	audioTransceiver  *webrtc.RTPTransceiver
	videoTransceiver  *webrtc.RTPTransceiver
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

func (webrtc_peer *WebRTCPeerCtx) ReplaceAudioTrack(track webrtc.TrackLocal) error {
	return webrtc_peer.audioTransceiver.Sender().ReplaceTrack(track)
}

func (webrtc_peer *WebRTCPeerCtx) ReplaceVideoTrack(track webrtc.TrackLocal) error {
	return webrtc_peer.videoTransceiver.Sender().ReplaceTrack(track)
}

func (webrtc_peer *WebRTCPeerCtx) Destroy() error {
	if webrtc_peer.connection == nil || webrtc_peer.connection.ConnectionState() != webrtc.PeerConnectionStateConnected {
		return nil
	}

	return webrtc_peer.connection.Close()
}
