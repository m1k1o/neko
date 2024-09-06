package types

import (
	"errors"

	"github.com/pion/webrtc/v3"
)

var (
	ErrWebRTCDataChannelNotFound = errors.New("webrtc data channel not found")
	ErrWebRTCConnectionNotFound  = errors.New("webrtc connection not found")
	ErrWebRTCStreamNotFound      = errors.New("webrtc stream not found")
)

type ICEServer struct {
	URLs       []string `mapstructure:"urls"       json:"urls"`
	Username   string   `mapstructure:"username"   json:"username,omitempty"`
	Credential string   `mapstructure:"credential" json:"credential,omitempty"`
}

type PeerVideo struct {
	Disabled bool   `json:"disabled"`
	ID       string `json:"id"`
	Video    string `json:"video"` // TODO: Remove this, used for compatibility with old clients.
	Auto     bool   `json:"auto"`
}

type PeerVideoRequest struct {
	Disabled *bool           `json:"disabled,omitempty"`
	Selector *StreamSelector `json:"selector,omitempty"`
	Auto     *bool           `json:"auto,omitempty"`
}

type PeerAudio struct {
	Disabled bool `json:"disabled"`
}

type PeerAudioRequest struct {
	Disabled *bool `json:"disabled,omitempty"`
}

type WebRTCPeer interface {
	CreateOffer(ICERestart bool) (*webrtc.SessionDescription, error)
	CreateAnswer() (*webrtc.SessionDescription, error)
	SetRemoteDescription(webrtc.SessionDescription) error
	SetCandidate(webrtc.ICECandidateInit) error

	SetPaused(isPaused bool) error
	Paused() bool

	SetVideo(PeerVideoRequest) error
	Video() PeerVideo
	SetAudio(PeerAudioRequest) error
	Audio() PeerAudio

	SendCursorPosition(x, y int) error
	SendCursorImage(cur *CursorImage, img []byte) error

	Destroy()
}

type WebRTCManager interface {
	Start()
	Shutdown() error

	ICEServers() []ICEServer

	CreatePeer(session Session) (*webrtc.SessionDescription, WebRTCPeer, error)
	SetCursorPosition(x, y int)
}
