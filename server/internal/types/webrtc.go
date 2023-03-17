package types

import (
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type Sample media.Sample

type WebRTCManager interface {
	Start()
	Shutdown() error
	CreatePeer(id string, session Session) (Peer, error)
	ICELite() bool
	ICEServers() []webrtc.ICEServer
	ImplicitControl() bool
}

type Peer interface {
	CreateOffer() (string, error)
	CreateAnswer() (string, error)
	SetOffer(sdp string) error
	SetAnswer(sdp string) error
	SetCandidate(candidateString string) error
	WriteData(v interface{}) error
	Destroy() error
}
