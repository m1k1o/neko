package types

import "github.com/pion/webrtc/v3"

type WebRTCPeer interface {
	SignalAnswer(sdp string) error
	SignalCandidate(candidate webrtc.ICECandidateInit) error

	Destroy() error
}

type WebRTCManager interface {
	Start()
	Shutdown() error

	ICELite() bool
	ICEServers() []string

	CreatePeer(session Session) (*webrtc.SessionDescription, error)
}
