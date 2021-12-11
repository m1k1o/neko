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
	Connected() bool
	Member() *Member
	SetMuted(muted bool)
	SetName(name string) error
	SetConnected(connected bool) error
	SetSocket(socket WebSocket) error
	SetPeer(peer Peer) error
	Address() string
	Kick(message string) error
	Send(v interface{}) error
	SignalLocalOffer(sdp string) error
	SignalLocalAnswer(sdp string) error
	SignalRemoteOffer(sdp string) error
	SignalRemoteAnswer(sdp string) error
	SignalCandidate(data string) error
}

type SessionManager interface {
	New(id string, admin bool, socket WebSocket) Session
	HasHost() bool
	IsHost(id string) bool
	SetHost(id string) error
	GetHost() (Session, bool)
	ClearHost()
	Has(id string) bool
	Get(id string) (Session, bool)
	SetControlLocked(locked bool)
	CanControl(id string) bool
	Members() []*Member
	Admins() []*Member
	Destroy(id string)
	Clear() error
	Broadcast(v interface{}, exclude interface{}) error
	AdminBroadcast(v interface{}, exclude interface{}) error
	OnHost(listener func(id string))
	OnHostCleared(listener func(id string))
	OnDestroy(listener func(id string, session Session))
	OnCreated(listener func(id string, session Session))
	OnConnected(listener func(id string, session Session))
}
