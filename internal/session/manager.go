package session

import (
	"fmt"
	"sync"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/config"
	"demodesk/neko/internal/utils"
)

func New(config *config.Session) *SessionManagerCtx {
	manager := &SessionManagerCtx{
		logger:    log.With().Str("module", "session").Logger(),
		host:      nil,
		hostMu:    sync.Mutex{},
		config:    config,
		members:   make(map[string]*SessionCtx),
		membersMu: sync.Mutex{},
		emmiter:   events.New(),
	}

	// create default admin account at startup
	_ = manager.Create("admin", types.MemberProfile{
		Secret: config.AdminPassword,
		Name: "Administrator",
		IsAdmin: true,
	})

	// create default user account at startup
	_ = manager.Create("user", types.MemberProfile{
		Secret: config.Password,
		Name: "User",
		IsAdmin: false,
	})

	return manager
}

type SessionManagerCtx struct {
	logger    zerolog.Logger
	host      types.Session
	hostMu    sync.Mutex
	config    *config.Session
	members   map[string]*SessionCtx
	membersMu sync.Mutex
	emmiter   events.EventEmmiter
}

func (manager *SessionManagerCtx) Create(id string, profile types.MemberProfile) types.Session {
	session := &SessionCtx{
		id:        id,
		manager:   manager,
		logger:    manager.logger.With().Str("id", id).Logger(),
		profile:   profile,
	}

	manager.membersMu.Lock()
	manager.members[id] = session
	manager.membersMu.Unlock()

	manager.emmiter.Emit("created", session)
	return session
}

func (manager *SessionManagerCtx) Get(id string) (types.Session, bool) {
	manager.membersMu.Lock()
	session, ok := manager.members[id]
	manager.membersMu.Unlock()

	return session, ok
}

func (manager *SessionManagerCtx) Delete(id string) error {
	manager.membersMu.Lock()
	session, ok := manager.members[id]
	if !ok {
		manager.membersMu.Unlock()
		return fmt.Errorf("Member not found.")
	}

	delete(manager.members, id)
	manager.membersMu.Unlock()

	var err error
	if session.IsConnected() {
		err = session.Disconnect("member deleted")
	}

	manager.emmiter.Emit("deleted", session)
	return err
}

// ---
// host
// ---

func (manager *SessionManagerCtx) HasHost() bool {
	manager.hostMu.Lock()
	defer manager.hostMu.Unlock()

	return manager.host != nil
}

func (manager *SessionManagerCtx) SetHost(host types.Session) {
	manager.hostMu.Lock()
	defer manager.hostMu.Unlock()

	manager.host = host
	manager.emmiter.Emit("host", host)
}

func (manager *SessionManagerCtx) GetHost() types.Session {
	manager.hostMu.Lock()
	defer manager.hostMu.Unlock()

	return manager.host
}

func (manager *SessionManagerCtx) ClearHost() {
	manager.hostMu.Lock()
	defer manager.hostMu.Unlock()

	host := manager.host
	manager.host = nil
	manager.emmiter.Emit("host_cleared", host)
}

// ---
// members list
// ---

func (manager *SessionManagerCtx) HasConnectedMembers() bool {
	manager.membersMu.Lock()
	defer manager.membersMu.Unlock()

	for _, session := range manager.members {
		if session.IsConnected() {
			return true
		}
	}

	return false
}

func (manager *SessionManagerCtx) Members() []types.Session {
	manager.membersMu.Lock()
	defer manager.membersMu.Unlock()

	var sessions []types.Session
	for _, session := range manager.members {
		sessions = append(sessions, session)
	}

	return sessions
}

func (manager *SessionManagerCtx) Broadcast(v interface{}, exclude interface{}) {
	manager.membersMu.Lock()
	defer manager.membersMu.Unlock()

	for id, session := range manager.members {
		if !session.IsConnected() {
			continue
		}

		if exclude != nil {
			if in, _ := utils.ArrayIn(id, exclude); in {
				continue
			}
		}

		if err := session.Send(v); err != nil {
			manager.logger.Warn().Err(err).Msgf("broadcasting event has failed")
		}
	}
}

func (manager *SessionManagerCtx) AdminBroadcast(v interface{}, exclude interface{}) {
	manager.membersMu.Lock()
	defer manager.membersMu.Unlock()

	for id, session := range manager.members {
		if !session.IsConnected() || !session.IsAdmin() {
			continue
		}

		if exclude != nil {
			if in, _ := utils.ArrayIn(id, exclude); in {
				continue
			}
		}

		if err := session.Send(v); err != nil {
			manager.logger.Warn().Err(err).Msgf("broadcasting admin event has failed")
		}
	}
}

// ---
// events
// ---

func (manager *SessionManagerCtx) OnHost(listener func(session types.Session)) {
	manager.emmiter.On("host", func(payload ...interface{}) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnHostCleared(listener func(session types.Session)) {
	manager.emmiter.On("host_cleared", func(payload ...interface{}) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnCreated(listener func(session types.Session)) {
	manager.emmiter.On("created", func(payload ...interface{}) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnDeleted(listener func(session types.Session)) {
	manager.emmiter.On("deleted", func(payload ...interface{}) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnConnected(listener func(session types.Session)) {
	manager.emmiter.On("connected", func(payload ...interface{}) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnDisconnected(listener func(session types.Session)) {
	manager.emmiter.On("disconnected", func(payload ...interface{}) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnProfileChanged(listener func(session types.Session)) {
	manager.emmiter.On("profile_changed", func(payload ...interface{}) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnStateChanged(listener func(session types.Session)) {
	manager.emmiter.On("state_changed", func(payload ...interface{}) {
		listener(payload[0].(*SessionCtx))
	})
}

// ---
// config
// ---

func (manager *SessionManagerCtx) ImplicitHosting() bool {
	return manager.config.ImplicitHosting
}
