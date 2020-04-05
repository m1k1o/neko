package types

type Sample struct {
	Data    []byte
	Samples uint32
}

type WebRTCManager interface {
	Start()
	Shutdown() error
	CreatePeer(id string, session Session) (string, bool, []string, error)
	ChangeScreenSize(width int, height int, rate int) error
}

type Peer interface {
	SignalAnswer(sdp string) error
	WriteData(v interface{}) error
	Destroy() error
}
