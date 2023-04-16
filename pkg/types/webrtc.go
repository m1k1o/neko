package types

import (
	"errors"

	"github.com/pion/webrtc/v3"
)

var (
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
	SetRemoteDescription(webrtc.SessionDescription) error
	SetCandidate(webrtc.ICECandidateInit) error

	SetVideoBitrate(bitrate int) error
	SetVideoID(videoID string) error
	GetVideoID() string
	SetPaused(isPaused bool) error
	SetVideoAuto(auto bool)
	VideoAuto() bool

	SendCursorPosition(x, y int) error
	SendCursorImage(cur *CursorImage, img []byte) error

	Destroy()
}

type WebRTCManager interface {
	Start()
	Shutdown() error

	ICEServers() []ICEServer

	CreatePeer(session Session, bitrate int, videoAuto bool) (*webrtc.SessionDescription, error)
	SetCursorPosition(x, y int)
}
