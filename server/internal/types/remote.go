package types

type RemoteManager interface {
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
	GetScreenSize() *ScreenSize
	ScreenConfigurations() map[int]ScreenConfiguration
	Move(x, y int)
	Scroll(x, y int)
	ButtonDown(code int) (*Button, error)
	KeyDown(code int) (*Key, error)
	ButtonUp(code int) (*Button, error)
	KeyUp(code int) (*Key, error)
	ReadClipboard() string
	WriteClipboard(data string)
	ResetKeys()
}
