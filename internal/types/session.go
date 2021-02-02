package types

import (
	"net/http"

	"github.com/pion/webrtc/v3"
)

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

type MemberState struct {
	IsConnected bool `json:"is_connected"`
	IsWatching  bool `json:"is_watching"`
}

type MembersDatabase interface {
	Connect() error
	Disconnect() error

	Insert(id string, profile MemberProfile) error	
	Update(id string, profile MemberProfile) error	
	Delete(id string) error	
	Select() (map[string]MemberProfile, error)
}

type Session interface {
	ID() string

	// profile
	VerifySecret(secret string) bool
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
	IsWatching() bool
	GetState() MemberState

	// websocket
	SetWebSocketPeer(websocket_peer WebSocketPeer)
	SetWebSocketConnected(connected bool)
	Send(v interface{}) error
	Disconnect(reason string) error

	// webrtc
	SetWebRTCPeer(webrtc_peer WebRTCPeer)
	SetWebRTCConnected(connected bool)
	SignalAnswer(sdp string) error
	SignalCandidate(candidate webrtc.ICECandidateInit) error
}

type SessionManager interface {
	Connect() error
	Disconnect() error

	Create(id string, profile MemberProfile) (Session, error)
	Update(id string, profile MemberProfile) error
	Get(id string) (Session, bool)
	Delete(id string) error

	SetHost(host Session)
	GetHost() Session
	ClearHost()

	HasConnectedMembers() bool
	Members() []Session
	Broadcast(v interface{}, exclude interface{})
	AdminBroadcast(v interface{}, exclude interface{})

	OnHost(listener func(session Session))
	OnHostCleared(listener func(session Session))
	OnCreated(listener func(session Session))
	OnDeleted(listener func(session Session))
	OnConnected(listener func(session Session))
	OnDisconnected(listener func(session Session))
	OnProfileChanged(listener func(session Session))
	OnStateChanged(listener func(session Session))

	ImplicitHosting() bool

	AuthenticateRequest(r *http.Request) (Session, error)
	Authenticate(id string, secret string) (Session, error)
}
