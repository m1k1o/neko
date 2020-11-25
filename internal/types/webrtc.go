package types

type WebRTCPeer interface {
	SignalAnswer(sdp string) error
	Destroy() error
}

type WebRTCManager interface {
	Start()
	Shutdown() error
	CreatePeer(session Session) (string, bool, []string, error)
}
