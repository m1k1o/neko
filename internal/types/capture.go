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

type CaptureManager interface {
	Start()
	Shutdown() error

	Broadcast() BroadcastManager
	Screencast() ScreencastManager

	VideoCodec() codec.RTP
	AudioCodec() codec.RTP

	OnVideoFrame(listener func(sample Sample))
	OnAudioFrame(listener func(sample Sample))

	StartStream()
	StopStream()
	Streaming() bool
}
