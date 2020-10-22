package webrtc

import (
	"sync"

	"github.com/pion/webrtc/v2"
)

type Peer struct {
	id            string
	api           *webrtc.API
	engine        *webrtc.MediaEngine
	manager       *WebRTCManager
	settings      *webrtc.SettingEngine
	connection    *webrtc.PeerConnection
	configuration *webrtc.Configuration
	mu            sync.Mutex
}

func (peer *Peer) SignalAnswer(sdp string) error {
	return peer.connection.SetRemoteDescription(webrtc.SessionDescription{SDP: sdp, Type: webrtc.SDPTypeAnswer})
}

func (peer *Peer) WriteData(v interface{}) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()
	return nil
}

func (peer *Peer) Destroy() error {
	if peer.connection != nil && peer.connection.ConnectionState() == webrtc.PeerConnectionStateConnected {
		if err := peer.connection.Close(); err != nil {
			return err
		}
	}

	return nil
}
