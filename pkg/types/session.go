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
	X int `json:"x"`
	Y int `json:"y"`
}

type SessionProfile struct {
	Id      string
	Token   string
	Profile MemberProfile
}

type SessionState struct {
	IsConnected bool `json:"is_connected"`
	IsWatching  bool `json:"is_watching"`
}

type Settings struct {
	PrivateMode       bool `json:"private_mode"`
	LockedControls    bool `json:"locked_controls"`
	ImplicitHosting   bool `json:"implicit_hosting"`
	InactiveCursors   bool `json:"inactive_cursors"`
	MercifulReconnect bool `json:"merciful_reconnect"`

	// plugin scope
	Plugins map[string]any `json:"plugins"`
}

type Session interface {
	ID() string
	Profile() MemberProfile
	State() SessionState
	IsHost() bool
	PrivateModeEnabled() bool

	// cursor
	SetCursor(cursor Cursor)

	// websocket
	ConnectWebSocketPeer(websocketPeer WebSocketPeer)
	DisconnectWebSocketPeer(websocketPeer WebSocketPeer, delayed bool)
	DestroyWebSocketPeer(reason string)
	Send(event string, payload any)

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
	GetHost() (Session, bool)
	ClearHost()

	SetCursor(cursor Cursor, session Session)
	PopCursors() map[Session][]Cursor

	Broadcast(event string, payload any, exclude ...string)
	AdminBroadcast(event string, payload any, exclude ...string)
	InactiveCursorsBroadcast(event string, payload any, exclude ...string)

	OnCreated(listener func(session Session))
	OnDeleted(listener func(session Session))
	OnConnected(listener func(session Session))
	OnDisconnected(listener func(session Session))
	OnProfileChanged(listener func(session Session))
	OnStateChanged(listener func(session Session))
	OnHostChanged(listener func(session Session))
	OnSettingsChanged(listener func(new Settings, old Settings))

	UpdateSettings(Settings)
	Settings() Settings
	CookieEnabled() bool

	CookieSetToken(w http.ResponseWriter, token string)
	CookieClearToken(w http.ResponseWriter, r *http.Request)
	Authenticate(r *http.Request) (Session, error)
}
