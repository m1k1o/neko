package types

import "net/http"

type Session interface {
	ID() string
	Name() string
	Admin() bool
	IsHost() bool
	Connected() bool
	SetName(name string)
	SetSocket(socket WebSocket)
	SetPeer(peer Peer)
	SetConnected(connected bool)
	Disconnect(reason string) error
	Send(v interface{}) error
	SignalAnswer(sdp string) error
}

type SessionManager interface {
	New(id string, admin bool) Session
	Get(id string) (Session, bool)
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
	OnCreated(listener func(session Session))
	OnConnected(listener func(session Session))
	OnDisconnected(listener func(session Session))

	// auth
	Authenticate(r *http.Request) (Session, error)
}
