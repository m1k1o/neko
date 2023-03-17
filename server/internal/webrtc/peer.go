package webrtc

import (
	"encoding/json"
	"sync"

	"github.com/pion/webrtc/v3"
)

type Peer struct {
	id         string
	mu         sync.Mutex
	manager    *WebRTCManager
	connection *webrtc.PeerConnection
}

func (peer *Peer) CreateOffer() (string, error) {
	desc, err := peer.connection.CreateOffer(nil)
	if err != nil {
		return "", err
	}

	err = peer.connection.SetLocalDescription(desc)
	if err != nil {
		return "", err
	}

	return desc.SDP, nil
}

func (peer *Peer) CreateAnswer() (string, error) {
	desc, err := peer.connection.CreateAnswer(nil)
	if err != nil {
		return "", err
	}

	err = peer.connection.SetLocalDescription(desc)
	if err != nil {
		return "", nil
	}

	return desc.SDP, nil
}

func (peer *Peer) SetOffer(sdp string) error {
	return peer.connection.SetRemoteDescription(webrtc.SessionDescription{SDP: sdp, Type: webrtc.SDPTypeOffer})
}

func (peer *Peer) SetAnswer(sdp string) error {
	return peer.connection.SetRemoteDescription(webrtc.SessionDescription{SDP: sdp, Type: webrtc.SDPTypeAnswer})
}

func (peer *Peer) SetCandidate(candidateString string) error {
	var candidate webrtc.ICECandidateInit
	err := json.Unmarshal([]byte(candidateString), &candidate)
	if err != nil {
		return err
	}

	return peer.connection.AddICECandidate(candidate)
}

func (peer *Peer) WriteData(v interface{}) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()
	return nil
}

func (peer *Peer) Destroy() error {
	if peer.connection != nil && peer.connection.ConnectionState() != webrtc.PeerConnectionStateClosed {
		if err := peer.connection.Close(); err != nil {
			return err
		}
	}

	return nil
}
