package types

import (
	"errors"

	"github.com/pion/webrtc/v3"
)

var (
	ErrWebRTCVideoNotFound       = errors.New("webrtc video not found")
	ErrWebRTCDataChannelNotFound = errors.New("webrtc data channel not found")
)

type ICEServer struct {
	URLs       []string `mapstructure:"urls"       json:"urls"`
	Username   string   `mapstructure:"username"   json:"username,omitempty"`
	Credential string   `mapstructure:"credential" json:"credential,omitempty"`
}

type WebRTCPeer interface {
	CreateOffer(ICERestart bool) (*webrtc.SessionDescription, error)
	SignalAnswer(sdp string) error
	SignalCandidate(candidate webrtc.ICECandidateInit) error

	SetVideoID(videoID string) error
	SendCursorPosition(x, y int) error
	SendCursorImage(cur *CursorImage, img []byte) error

	Destroy()
}

type WebRTCManager interface {
	Start()
	Shutdown() error

	ICEServers() []ICEServer

	CreatePeer(session Session, videoID string) (*webrtc.SessionDescription, error)
}
