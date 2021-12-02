package session

import (
	"fmt"
	"sync"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/utils"
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
	mu      sync.Mutex
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

	manager.mu.Lock()
	manager.members[id] = session
	if !manager.remote.Streaming() && len(manager.members) > 0 {
		manager.remote.StartStream()
	}
	manager.mu.Unlock()

	manager.emmiter.Emit("created", id, session)
	return session
}

func (manager *SessionManager) HasHost() bool {
	return manager.host != ""
}

func (manager *SessionManager) IsHost(id string) bool {
	return manager.host == id
}

func (manager *SessionManager) SetHost(id string) error {
	manager.mu.Lock()
	_, ok := manager.members[id]
	manager.mu.Unlock()

	if ok {
		manager.host = id
		manager.emmiter.Emit("host", id)
		return nil
	}

	return fmt.Errorf("invalid session id %s", id)
}

func (manager *SessionManager) GetHost() (types.Session, bool) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	host, ok := manager.members[manager.host]
	return host, ok
}

func (manager *SessionManager) ClearHost() {
	id := manager.host
	manager.host = ""
	manager.emmiter.Emit("host_cleared", id)
}

func (manager *SessionManager) Has(id string) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	_, ok := manager.members[id]
	return ok
}

func (manager *SessionManager) Get(id string) (types.Session, bool) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	session, ok := manager.members[id]
	return session, ok
}

func (manager *SessionManager) Admins() []*types.Member {
	manager.mu.Lock()
	defer manager.mu.Unlock()

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
	manager.mu.Lock()
	defer manager.mu.Unlock()

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

func (manager *SessionManager) Destroy(id string) {
	manager.mu.Lock()
	session, ok := manager.members[id]
	if ok {
		err := session.destroy()
		delete(manager.members, id)

		if manager.remote.Streaming() && len(manager.members) <= 0 {
			manager.remote.StopStream()
		}
		manager.mu.Unlock()

		manager.emmiter.Emit("destroyed", id, session)
		manager.logger.Err(err).Str("session_id", id).Msg("destroying session")
		return
	}

	manager.mu.Unlock()
}

func (manager *SessionManager) Clear() error {
	return nil
}

func (manager *SessionManager) Broadcast(v interface{}, exclude interface{}) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

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
