package types

import (
	"errors"
	"net/http"
)

var (
	ErrSessionNotFound         = errors.New("session not found")
	ErrSessionAlreadyExists    = errors.New("session already exists")
	ErrSessionAlreadyConnected = errors.New("session is already connected")
	ErrSessionLoginDisabled    = errors.New("session login disabled")
)

type Cursor struct {
	X int
	Y int
}

type SessionState struct {
	IsConnected bool `json:"is_connected"`
	IsWatching  bool `json:"is_watching"`
}

type Session interface {
	ID() string
	Profile() MemberProfile
	State() SessionState
	IsHost() bool

	// cursor
	SetCursor(x, y int)

	// websocket
	SetWebSocketPeer(websocketPeer WebSocketPeer)
	SetWebSocketConnected(websocketPeer WebSocketPeer, connected bool)
	GetWebSocketPeer() WebSocketPeer
	Send(event string, payload interface{})

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

	SetCursor(x, y int, session Session)
	PopCursors() map[Session]Cursor

	Broadcast(event string, payload interface{}, exclude interface{})
	AdminBroadcast(event string, payload interface{}, exclude interface{})

	OnCreated(listener func(session Session))
	OnDeleted(listener func(session Session))
	OnConnected(listener func(session Session))
	OnDisconnected(listener func(session Session))
	OnProfileChanged(listener func(session Session))
	OnStateChanged(listener func(session Session))
	OnHostChanged(listener func(session Session))

	ImplicitHosting() bool
	CookieEnabled() bool
	MercifulReconnect() bool

	CookieSetToken(w http.ResponseWriter, token string)
	CookieClearToken(w http.ResponseWriter, r *http.Request)
	Authenticate(r *http.Request) (Session, error)
}
