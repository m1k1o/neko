package session

import (
	"fmt"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"n.eko.moe/neko/internal/types"
	"n.eko.moe/neko/internal/utils"
)

func New(remote types.RemoteManager) *SessionManager {
	return &SessionManager{
		logger:  log.With().Str("module", "session").Logger(),
		host:    "",
		remote:  remote,
		members: make(map[string]*Session),
		emmiter: events.New(),
	}
}

type SessionManager struct {
	logger  zerolog.Logger
	host    string
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
	manager.emmiter.Emit("created", id, session)

	if manager.remote.Streaming() != true && len(manager.members) > 0 {
		manager.remote.StartStream()
	}

	return session
}

func (manager *SessionManager) HasHost() bool {
	return manager.host != ""
}

func (manager *SessionManager) IsHost(id string) bool {
	return manager.host == id
}

func (manager *SessionManager) SetHost(id string) error {
	_, ok := manager.members[id]
	if ok {
		manager.host = id
		manager.emmiter.Emit("host", id)
		return nil
	}
	return fmt.Errorf("invalid session id %s", id)
}

func (manager *SessionManager) GetHost() (types.Session, bool) {
	host, ok := manager.members[manager.host]
	return host, ok
}

func (manager *SessionManager) ClearHost() {
	id := manager.host
	manager.host = ""
	manager.emmiter.Emit("host_cleared", id)
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
		err := session.destroy()
		delete(manager.members, id)

		if manager.remote.Streaming() != false && len(manager.members) <= 0 {
			manager.remote.StopStream()
		}

		manager.emmiter.Emit("destroyed", id, session)
		return err
	}

	return nil
}

func (manager *SessionManager) Clear() error {
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

func (manager *SessionManager) OnHost(listener func(id string)) {
	manager.emmiter.On("host", func(payload ...interface{}) {
		listener(payload[0].(string))
	})
}

func (manager *SessionManager) OnHostCleared(listener func(id string)) {
	manager.emmiter.On("host_cleared", func(payload ...interface{}) {
		listener(payload[0].(string))
	})
}

func (manager *SessionManager) OnDestroy(listener func(id string, session types.Session)) {
	manager.emmiter.On("destroyed", func(payload ...interface{}) {
		listener(payload[0].(string), payload[1].(*Session))
	})
}

func (manager *SessionManager) OnCreated(listener func(id string, session types.Session)) {
	manager.emmiter.On("created", func(payload ...interface{}) {
		listener(payload[0].(string), payload[1].(*Session))
	})
}

func (manager *SessionManager) OnConnected(listener func(id string, session types.Session)) {
	manager.emmiter.On("connected", func(payload ...interface{}) {
		listener(payload[0].(string), payload[1].(*Session))
	})
}
