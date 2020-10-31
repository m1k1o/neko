package types

type Member struct {
	ID    string `json:"id"`
	Name  string `json:"displayname"`
	Admin bool   `json:"admin"`
	Muted bool   `json:"muted"`
}

type Session interface {
	ID() string
	Name() string
	Admin() bool
	Muted() bool
	IsHost() bool
	Connected() bool
	Member() *Member
	SetMuted(muted bool)
	SetName(name string)
	SetConnected()
	SetSocket(socket WebSocket)
	SetPeer(peer Peer)
	Address() string
	Disconnect(message string) error
	Send(v interface{}) error
	SignalAnswer(sdp string) error
}

type SessionManager interface {
	New(id string, admin bool, socket WebSocket) Session
	HasHost() bool
	SetHost(Session)
	GetHost() Session
	ClearHost()
	Has(id string) bool
	Get(id string) (Session, bool)
	Members() []*Member
	Admins() []*Member
	Destroy(id string) error
	Broadcast(v interface{}, exclude interface{}) error
	OnHost(listener func(session Session))
	OnHostCleared(listener func(session Session))
	OnDestroy(listener func(id string))
	OnCreated(listener func(session Session))
	OnConnected(listener func(session Session))
}
