package types

type Sample struct {
	Data    []byte
	Samples uint32
}

type WebRTCManager interface {
	Start()
	Shutdown() error
	CreatePeer(id string, sdp string) (string, Peer, error)
}

type Peer interface {
	WriteVideoSample(sample Sample) error
	WriteAudioSample(sample Sample) error
	WriteData(v interface{}) error
	Destroy() error
}
