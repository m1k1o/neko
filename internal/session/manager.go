package session

import (
	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

func New(remote types.RemoteManager) *SessionManager {
	return &SessionManager{
		logger:  log.With().Str("module", "session").Logger(),
		host:    nil,
		remote:  remote,
		members: make(map[string]*Session),
		emmiter: events.New(),
	}
}

type SessionManager struct {
	logger  zerolog.Logger
	host    types.Session
	remote  types.RemoteManager
	members map[string]*Session
	emmiter events.EventEmmiter
}

func (manager *SessionManager) New(id string, admin bool, socket types.WebSocket) types.Session {
	session := &Session{
		id:        id,
		admin:     admin,
		manager:   manager,
		socket:    socket,
		logger:    manager.logger.With().Str("id", id).Logger(),
		connected: false,
	}

	manager.members[id] = session
	manager.emmiter.Emit("created", session)

	if !manager.remote.Streaming() && len(manager.members) > 0 {
		manager.remote.StartStream()
	}

	return session
}

func (manager *SessionManager) HasHost() bool {
	return manager.host != nil
}

func (manager *SessionManager) SetHost(host types.Session) {
	manager.host = host
	manager.emmiter.Emit("host", host)
}

func (manager *SessionManager) GetHost() types.Session {
	return manager.host
}

func (manager *SessionManager) ClearHost() {
	host := manager.host
	manager.host = nil
	manager.emmiter.Emit("host_cleared", host)
}

func (manager *SessionManager) Has(id string) bool {
	_, ok := manager.members[id]
	return ok
}

func (manager *SessionManager) Get(id string) (types.Session, bool) {
	session, ok := manager.members[id]
	return session, ok
}

func (manager *SessionManager) Admins() []*types.Member {
	members := []*types.Member{}
	for _, session := range manager.members {
		if !session.connected || !session.admin {
			continue
		}

		member := session.Member()
		if member != nil {
			members = append(members, member)
		}
	}
	return members
}

func (manager *SessionManager) Members() []*types.Member {
	members := []*types.Member{}
	for _, session := range manager.members {
		if !session.connected {
			continue
		}

		member := session.Member()
		if member != nil {
			members = append(members, member)
		}
	}
	return members
}

func (manager *SessionManager) Destroy(id string) error {
	session, ok := manager.members[id]
	if ok {
		delete(manager.members, id)
		err := session.destroy()

		if !manager.remote.Streaming() && len(manager.members) <= 0 {
			manager.remote.StopStream()
		}

		manager.emmiter.Emit("destroy", id)
		return err
	}

	return nil
}

func (manager *SessionManager) Broadcast(v interface{}, exclude interface{}) error {
	for id, session := range manager.members {
		if !session.connected {
			continue
		}

		if exclude != nil {
			if in, _ := utils.ArrayIn(id, exclude); in {
				continue
			}
		}

		if err := session.Send(v); err != nil {
			return err
		}
	}
	return nil
}

func (manager *SessionManager) OnHost(listener func(session types.Session)) {
	manager.emmiter.On("host", func(payload ...interface{}) {
		listener(payload[0].(*Session))
	})
}

func (manager *SessionManager) OnHostCleared(listener func(session types.Session)) {
	manager.emmiter.On("host_cleared", func(payload ...interface{}) {
		listener(payload[0].(*Session))
	})
}

func (manager *SessionManager) OnDestroy(listener func(id string)) {
	manager.emmiter.On("destroy", func(payload ...interface{}) {
		listener(payload[0].(string))
	})
}

func (manager *SessionManager) OnCreated(listener func(session types.Session)) {
	manager.emmiter.On("created", func(payload ...interface{}) {
		listener(payload[0].(*Session))
	})
}

func (manager *SessionManager) OnConnected(listener func(session types.Session)) {
	manager.emmiter.On("connected", func(payload ...interface{}) {
		listener(payload[0].(*Session))
	})
}
