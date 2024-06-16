package types

import (
	"errors"
	"net/http"
	"time"
)

var (
	ErrSessionNotFound         = errors.New("session not found")
	ErrSessionAlreadyExists    = errors.New("session already exists")
	ErrSessionAlreadyConnected = errors.New("session is already connected")
	ErrSessionLoginDisabled    = errors.New("session login disabled")
	ErrSessionLoginsLocked     = errors.New("session logins locked")
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
	// when the session was last connected
	ConnectedSince *time.Time `json:"connected_since,omitempty"`
	// when the session was last not connected
	NotConnectedSince *time.Time `json:"not_connected_since,omitempty"`

	IsWatching bool `json:"is_watching"`
	// when the session was last watching
	WatchingSince *time.Time `json:"watching_since,omitempty"`
	// when the session was last not watching
	NotWatchingSince *time.Time `json:"not_watching_since,omitempty"`
}

type Settings struct {
	PrivateMode       bool `json:"private_mode"`
	LockedLogins      bool `json:"locked_logins"`
	LockedControls    bool `json:"locked_controls"`
	ControlProtection bool `json:"control_protection"`
	ImplicitHosting   bool `json:"implicit_hosting"`
	InactiveCursors   bool `json:"inactive_cursors"`
	MercifulReconnect bool `json:"merciful_reconnect"`

	// plugin scope
	Plugins PluginSettings `json:"plugins"`
}

type Session interface {
	ID() string
	Profile() MemberProfile
	State() SessionState
	IsHost() bool
	SetAsHost()
	SetAsHostBy(session Session)
	ClearHost()
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
	Disconnect(id string) error
	Get(id string) (Session, bool)
	GetByToken(token string) (Session, bool)
	List() []Session
	Range(func(Session) bool)

	GetHost() (Session, bool)

	SetCursor(cursor Cursor, session Session)
	PopCursors() map[Session][]Cursor

	Broadcast(event string, payload any, exclude ...string)
	AdminBroadcast(event string, payload any, exclude ...string)
	InactiveCursorsBroadcast(event string, payload any, exclude ...string)

	OnCreated(listener func(session Session))
	OnDeleted(listener func(session Session))
	OnConnected(listener func(session Session))
	OnDisconnected(listener func(session Session))
	OnProfileChanged(listener func(session Session, new, old MemberProfile))
	OnStateChanged(listener func(session Session))
	OnHostChanged(listener func(session, host Session))
	OnSettingsChanged(listener func(session Session, new, old Settings))

	UpdateSettingsFunc(session Session, f func(settings *Settings) bool)
	Settings() Settings
	CookieEnabled() bool

	CookieSetToken(w http.ResponseWriter, token string)
	CookieClearToken(w http.ResponseWriter, r *http.Request)
	Authenticate(r *http.Request) (Session, error)
}
