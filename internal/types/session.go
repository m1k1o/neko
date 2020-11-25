package types

import "net/http"

type MemberProfile struct {
	//Token            string
	Name             string
	IsAdmin         bool
	//Enabled          bool
	//CanControl      bool
	//CanWatch        bool
	//ClipboardAccess bool
}

type Session interface {
	ID() string
	Name() string
	Admin() bool
	IsHost() bool
	Connected() bool
	SetName(name string)
	SetWebSocketPeer(websocket_peer WebSocketPeer)
	SetWebSocketConnected(connected bool)
	SetWebRTCPeer(webrtc_peer WebRTCPeer)
	SetWebRTCConnected(connected bool)
	Disconnect(reason string) error
	Send(v interface{}) error
	SignalAnswer(sdp string) error
}

type SessionManager interface {
	Create(profile MemberProfile) (Session, error)
	Get(id string) (Session, bool)
	Delete(id string) error

	HasHost() bool
	SetHost(host Session)
	GetHost() Session
	ClearHost()

	Admins() []Session
	Members() []Session
	Broadcast(v interface{}, exclude interface{})
	AdminBroadcast(v interface{}, exclude interface{})

	OnHost(listener func(session Session))
	OnHostCleared(listener func(session Session))
	OnCreated(listener func(session Session))
	OnConnected(listener func(session Session))
	OnDisconnected(listener func(session Session))

	// auth
	Authenticate(r *http.Request) (Session, error)
}
