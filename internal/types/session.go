package types

import "net/http"

type Session interface {
	ID() string
	Name() string
	Admin() bool
	Muted() bool
	IsHost() bool
	Connected() bool
	Address() string
	SetMuted(muted bool)
	SetName(name string)
	SetSocket(socket WebSocket)
	SetPeer(peer Peer)
	SetConnected()
	Disconnect(reason string) error
	Send(v interface{}) error
	SignalAnswer(sdp string) error
}

type SessionManager interface {
	New(id string, admin bool, socket WebSocket) Session
	Get(id string) (Session, bool)
	Has(id string) bool
	Destroy(id string) error

	HasHost() bool
	SetHost(host Session)
	GetHost() Session
	ClearHost()

	Admins() []Session
	Members() []Session
	Broadcast(v interface{}, exclude interface{}) error

	OnHost(listener func(session Session))
	OnHostCleared(listener func(session Session))
	OnDestroy(listener func(id string))
	OnCreated(listener func(session Session))
	OnConnected(listener func(session Session))

	// auth
	Authenticate(r *http.Request) (string, string, bool, error)
}
