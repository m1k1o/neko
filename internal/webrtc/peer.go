package webrtc

import (
	"github.com/pion/webrtc/v2"
)

type PeerCtx struct {
	api           *webrtc.API
	engine        *webrtc.MediaEngine
	settings      *webrtc.SettingEngine
	connection    *webrtc.PeerConnection
	configuration *webrtc.Configuration
}

func (peer *PeerCtx) SignalAnswer(sdp string) error {
	return peer.connection.SetRemoteDescription(webrtc.SessionDescription{
		SDP: sdp,
		Type: webrtc.SDPTypeAnswer,
	})
}

func (peer *PeerCtx) Destroy() error {
	if peer.connection == nil || peer.connection.ConnectionState() != webrtc.PeerConnectionStateConnected {
		return nil
	}
	
	return peer.connection.Close()
}
