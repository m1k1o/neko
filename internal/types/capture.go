package types

type Sample struct {
	Data    []byte
	Samples uint32
}

type BroadcastManager interface {
	Start(url string) error
	Stop()
	Enabled() bool
	Url() string
}

type CaptureManager interface {
	Start()
	Shutdown() error

	Broadcast() BroadcastManager

	VideoCodec() string
	AudioCodec() string

	OnVideoFrame(listener func(sample Sample))
	OnAudioFrame(listener func(sample Sample))

	StartStream()
	StopStream()
	Streaming() bool
}
