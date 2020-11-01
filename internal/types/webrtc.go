package types

type WebRTCManager interface {
	Start()
	Shutdown() error
	CreatePeer(session Session) (string, bool, []string, error)
}

type Peer interface {
	SignalAnswer(sdp string) error
	Destroy() error
}
