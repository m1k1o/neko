package types

import (
	"time"

	"github.com/pion/webrtc/v3"
)

type Sample struct {
	Data      []byte
	Timestamp time.Time
	Duration  time.Duration
}

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
