package types

type Sample struct {
	Data    []byte
	Samples uint32
}

type CaptureManager interface {
	Start()
	Shutdown() error

	VideoCodec() string
	AudioCodec() string

	OnVideoFrame(listener func(sample Sample))
	OnAudioFrame(listener func(sample Sample))

	StartStream()
	StopStream()
	Streaming() bool

	ChangeResolution(width int, height int, rate int) error

	// broacast
	StartBroadcast(url string)
	StopBroadcast()
	IsBoradcasting() bool
	BroadcastUrl() string
}
