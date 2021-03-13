package types

import "net/http"

type MemberProfile struct {
	Secret             string `json:"secret,omitempty"`
	Name               string `json:"name"`
	IsAdmin            bool   `json:"is_admin"`
	CanLogin           bool   `json:"can_login"`
	CanConnect         bool   `json:"can_connect"`
	CanWatch           bool   `json:"can_watch"`
	CanHost            bool   `json:"can_host"`
	CanAccessClipboard bool   `json:"can_access_clipboard"`
}

type SessionState struct {
	IsConnected bool `json:"is_connected"`
	IsWatching  bool `json:"is_watching"`
}

type Session interface {
	ID() string

	// profile
	Name() string
	IsAdmin() bool
	CanLogin() bool
	CanConnect() bool
	CanWatch() bool
	CanHost() bool
	CanAccessClipboard() bool
	GetProfile() MemberProfile

	// state
	IsHost() bool
	IsConnected() bool
	GetState() SessionState

	// websocket
	SetWebSocketPeer(websocketPeer WebSocketPeer)
	SetWebSocketConnected(connected bool)
	Send(v interface{}) error
	Disconnect(reason string) error

	// webrtc
	SetWebRTCPeer(webrtcPeer WebRTCPeer)
	SetWebRTCConnected(webrtcPeer WebRTCPeer, connected bool)
	GetWebRTCPeer() WebRTCPeer
}

type SessionManager interface {
	Create(id string, profile MemberProfile) (Session, string, error)
	Update(id string, profile MemberProfile) error
	Delete(id string) error
	Get(id string) (Session, bool)
	GetByToken(token string) (Session, bool)
	List() []Session

	SetHost(host Session)
	GetHost() Session
	ClearHost()

	Broadcast(v interface{}, exclude interface{})
	AdminBroadcast(v interface{}, exclude interface{})

	OnCreated(listener func(session Session))
	OnDeleted(listener func(session Session))
	OnConnected(listener func(session Session))
	OnDisconnected(listener func(session Session))
	OnProfileChanged(listener func(session Session))
	OnStateChanged(listener func(session Session))
	OnHostChanged(listener func(session Session))

	ImplicitHosting() bool

	Authenticate(r *http.Request) (Session, error)
}
