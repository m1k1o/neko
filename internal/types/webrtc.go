package types

import "github.com/pion/webrtc/v3"

type ICEServer struct {
	URLs       []string `mapstructure:"urls"       json:"urls"`
	Username   string   `mapstructure:"username"   json:"username,omitempty"`
	Credential string   `mapstructure:"credential" json:"credential,omitempty"`
}

type WebRTCPeer interface {
	CreateOffer(ICETrickle bool, ICERestart bool) (*webrtc.SessionDescription, error)
	SignalAnswer(sdp string) error
	SignalCandidate(candidate webrtc.ICECandidateInit) error

	SetVideoID(videoID string) error
	SendCursorPosition(x, y int) error
	SendCursorImage(cur *CursorImage, img []byte) error

	Destroy() error
}

type WebRTCManager interface {
	Start()
	Shutdown() error

	ICEServers() []ICEServer

	CreatePeer(session Session, videoID string) (*webrtc.SessionDescription, error)
}
