package types

type CaptureManager interface {
	VideoCodec() string
	AudioCodec() string
	Start()
	Shutdown() error
	OnVideoFrame(listener func(sample Sample))
	OnAudioFrame(listener func(sample Sample))
	StartStream()
	StopStream()
	Streaming() bool
	ChangeResolution(width int, height int, rate int) error
}
