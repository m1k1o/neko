package session

import (
	"fmt"
	"sync"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

func New(config *config.Session) *SessionManagerCtx {
	return &SessionManagerCtx{
		logger:     log.With().Str("module", "session").Logger(),
		config:     config,
		host:       nil,
		hostMu:     sync.Mutex{},
		tokens:     make(map[string]string),
		sessions:   make(map[string]*SessionCtx),
		sessionsMu: sync.Mutex{},
		emmiter:    events.New(),
	}
}

type SessionManagerCtx struct {
	logger     zerolog.Logger
	config     *config.Session
	host       types.Session
	hostMu     sync.Mutex
	tokens     map[string]string
	sessions   map[string]*SessionCtx
	sessionsMu sync.Mutex
	emmiter    events.EventEmmiter
}

func (manager *SessionManagerCtx) Create(id string, profile types.MemberProfile) (types.Session, string, error) {
	token, err := utils.NewUID(64)
	if err != nil {
		return nil, "", err
	}

	manager.sessionsMu.Lock()
	if _, ok := manager.sessions[id]; ok {
		manager.sessionsMu.Unlock()
		return nil, "", fmt.Errorf("Session id already exists.")
	}

	if _, ok := manager.tokens[token]; ok {
		manager.sessionsMu.Unlock()
		return nil, "", fmt.Errorf("Session token already exists.")
	}

	session := &SessionCtx{
		id:      id,
		token:   token,
		manager: manager,
		logger:  manager.logger.With().Str("id", id).Logger(),
		profile: profile,
	}

	manager.tokens[token] = id
	manager.sessions[id] = session
	manager.sessionsMu.Unlock()

	manager.emmiter.Emit("created", session)
	return session, token, nil
}

func (manager *SessionManagerCtx) Update(id string, profile types.MemberProfile) error {
	manager.sessionsMu.Lock()

	session, ok := manager.sessions[id]
	if !ok {
		manager.sessionsMu.Unlock()
		return fmt.Errorf("Session id not found.")
	}

	session.profile = profile
	manager.sessionsMu.Unlock()

	manager.emmiter.Emit("profile_changed", session)
	session.profileChanged()
	return nil
}

func (manager *SessionManagerCtx) Delete(id string) error {
	manager.sessionsMu.Lock()
	session, ok := manager.sessions[id]
	if !ok {
		manager.sessionsMu.Unlock()
		return fmt.Errorf("Session id not found.")
	}

	if _, ok := manager.tokens[session.token]; ok {
		delete(manager.tokens, session.token)
	}

	delete(manager.sessions, id)
	manager.sessionsMu.Unlock()

	var err error
	if session.IsConnected() {
		err = session.Disconnect("session deleted")
	}

	manager.emmiter.Emit("deleted", session)
	return err
}

func (manager *SessionManagerCtx) Get(id string) (types.Session, bool) {
	manager.sessionsMu.Lock()
	defer manager.sessionsMu.Unlock()

	session, ok := manager.sessions[id]
	return session, ok
}

func (manager *SessionManagerCtx) GetByToken(token string) (types.Session, bool) {
	manager.sessionsMu.Lock()
	id, ok := manager.tokens[token]
	manager.sessionsMu.Unlock()

	if !ok {
		return nil, false
	}

	return manager.Get(id)
}

func (manager *SessionManagerCtx) List() []types.Session {
	manager.sessionsMu.Lock()
	defer manager.sessionsMu.Unlock()

	var sessions []types.Session
	for _, session := range manager.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// ---
// host
// ---

func (manager *SessionManagerCtx) SetHost(host types.Session) {
	manager.hostMu.Lock()
	manager.host = host
	manager.hostMu.Unlock()

	manager.emmiter.Emit("host_changed", host)
}

func (manager *SessionManagerCtx) GetHost() types.Session {
	manager.hostMu.Lock()
	defer manager.hostMu.Unlock()

	return manager.host
}

func (manager *SessionManagerCtx) ClearHost() {
	manager.SetHost(nil)
}

// ---
// broadcasts
// ---

func (manager *SessionManagerCtx) Broadcast(v interface{}, exclude interface{}) {
	manager.sessionsMu.Lock()
	defer manager.sessionsMu.Unlock()

	for id, session := range manager.sessions {
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
	manager.sessionsMu.Lock()
	defer manager.sessionsMu.Unlock()

	for id, session := range manager.sessions {
		if !session.IsConnected() || !session.Profile().IsAdmin {
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

func (manager *SessionManagerCtx) OnHostChanged(listener func(session types.Session)) {
	manager.emmiter.On("host_changed", func(payload ...interface{}) {
		if payload[0] == nil {
			listener(nil)
		} else {
			listener(payload[0].(*SessionCtx))
		}
	})
}

// ---
// config
// ---

func (manager *SessionManagerCtx) ImplicitHosting() bool {
	return manager.config.ImplicitHosting
}
