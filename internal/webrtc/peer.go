package webrtc

import "github.com/pion/webrtc/v3"

type WebRTCPeerCtx struct {
	api         *webrtc.API
	connection  *webrtc.PeerConnection
	dataChannel *webrtc.DataChannel
	changeVideo func(videoID string) error
}

func (peer *WebRTCPeerCtx) SignalAnswer(sdp string) error {
	return peer.connection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  sdp,
		Type: webrtc.SDPTypeAnswer,
	})
}

func (peer *WebRTCPeerCtx) SignalCandidate(candidate webrtc.ICECandidateInit) error {
	return peer.connection.AddICECandidate(candidate)
}

func (peer *WebRTCPeerCtx) SetVideoID(videoID string) error {
	return peer.changeVideo(videoID)
}

func (peer *WebRTCPeerCtx) Destroy() error {
	if peer.connection == nil || peer.connection.ConnectionState() != webrtc.PeerConnectionStateConnected {
		return nil
	}

	return peer.connection.Close()
}
