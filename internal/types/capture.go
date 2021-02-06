package types

import (
	"github.com/pion/webrtc/v3/pkg/media"

	"demodesk/neko/internal/types/codec"
)

type Sample media.Sample

type BroadcastManager interface {
	Start(url string) error
	Stop()
	Started() bool
	Url() string
}

type ScreencastManager interface {
	Enabled() bool
	Started() bool
	Image() ([]byte, error)
}

type StreamManager interface {
	Codec() codec.RTPCodec

	AddListener(listener *func(sample Sample))
	RemoveListener(listener *func(sample Sample))
	ListenersCount() int

	Start() error
	Stop()
	Started() bool
}

type CaptureManager interface {
	Start()
	Shutdown() error

	Broadcast() BroadcastManager
	Screencast() ScreencastManager
	Audio() StreamManager
	Video(videoID string) (StreamManager, bool)
	VideoIDs() []string
}
