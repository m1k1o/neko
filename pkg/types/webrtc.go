package types

import (
	"errors"

	"github.com/pion/webrtc/v3"
)

var (
	ErrWebRTCVideoNotFound       = errors.New("webrtc video not found")
	ErrWebRTCDataChannelNotFound = errors.New("webrtc data channel not found")
	ErrWebRTCConnectionNotFound  = errors.New("webrtc connection not found")
)

type ICEServer struct {
	URLs       []string `mapstructure:"urls"       json:"urls"`
	Username   string   `mapstructure:"username"   json:"username,omitempty"`
	Credential string   `mapstructure:"credential" json:"credential,omitempty"`
}

type WebRTCPeer interface {
	CreateOffer(ICERestart bool) (*webrtc.SessionDescription, error)
	CreateAnswer() (*webrtc.SessionDescription, error)
	SetOffer(sdp string) error
	SetAnswer(sdp string) error
	SetCandidate(candidate webrtc.ICECandidateInit) error

	SetVideoID(videoID string) error
	SetPaused(isPaused bool) error

	SendCursorPosition(x, y int) error
	SendCursorImage(cur *CursorImage, img []byte) error

	Destroy()
}

type WebRTCManager interface {
	Start()
	Shutdown() error

	ICEServers() []ICEServer

	CreatePeer(session Session, videoID string) (*webrtc.SessionDescription, error)
	SetCursorPosition(x, y int)
}
