package types

import (
	"github.com/pion/webrtc/v3/pkg/media"

	"demodesk/neko/internal/types/codec"
)

type Sample media.Sample

type BroadcastManager interface {
	Start(url string) error
	Stop()
	Enabled() bool
	Url() string
}

type ScreencastManager interface {
	Enabled() bool
	Started() bool
	Image() ([]byte, error)
}

type StreamManager interface {
	Shutdown()

	Codec() codec.RTPCodec
	OnSample(listener func(sample Sample))

	Start()
	Stop()
	Enabled() bool
}

type CaptureManager interface {
	Start()
	Shutdown() error

	Broadcast() BroadcastManager
	Screencast() ScreencastManager
	Audio() StreamManager
	Video() StreamManager

	StartStream()
	StopStream()
	Streaming() bool
}
