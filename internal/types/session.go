package types

import "net/http"

type MemberProfile struct {
	Secret             string
	Name               string

	IsAdmin            bool
	CanLogin           bool
	CanConnect         bool
	CanWatch           bool
	CanHost            bool
	CanAccessClipboard bool
}

type Session interface {
	ID() string

	VerifySecret(secret string) bool
	Name() string
	IsAdmin() bool
	CanLogin() bool
	CanConnect() bool
	CanWatch() bool
	CanHost() bool
	CanAccessClipboard() bool
	SetProfile(profile MemberProfile)

	IsHost() bool
	IsConnected() bool
	IsReceiving() bool
	Disconnect(reason string) error

	SetWebSocketPeer(websocket_peer WebSocketPeer)
	SetWebSocketConnected(connected bool)
	Send(v interface{}) error

	SetWebRTCPeer(webrtc_peer WebRTCPeer)
	SetWebRTCConnected(connected bool)
	SignalAnswer(sdp string) error
}

type SessionManager interface {
	Create(id string, profile MemberProfile) Session
	Get(id string) (Session, bool)
	Delete(id string) error

	HasHost() bool
	SetHost(host Session)
	GetHost() Session
	ClearHost()

	Members() []Session
	Broadcast(v interface{}, exclude interface{})
	AdminBroadcast(v interface{}, exclude interface{})

	OnHost(listener func(session Session))
	OnHostCleared(listener func(session Session))
	OnConnected(listener func(session Session))
	OnDisconnected(listener func(session Session))

	ImplicitHosting() bool

	Authenticate(r *http.Request) (Session, error)
}
