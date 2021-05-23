package types

import (
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type Sample media.Sample

type WebRTCManager interface {
	Start()
	Shutdown() error
	CreatePeer(id string, session Session) (string, bool, []webrtc.ICEServer, error)
}

type Peer interface {
	SignalAnswer(sdp string) error
	WriteData(v interface{}) error
	Destroy() error
}
