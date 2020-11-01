package session

import (
	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/config"
	"demodesk/neko/internal/utils"
)

func New(capture types.CaptureManager, config *config.Session) *SessionManagerCtx {
	return &SessionManagerCtx{
		logger:  log.With().Str("module", "session").Logger(),
		host:    nil,
		capture: capture,
		config:  config,
		members: make(map[string]*SessionCtx),
		emmiter: events.New(),
	}
}

type SessionManagerCtx struct {
	logger  zerolog.Logger
	host    types.Session
	capture types.CaptureManager
	config  *config.Session
	members map[string]*SessionCtx
	emmiter events.EventEmmiter
}

func (manager *SessionManagerCtx) New(id string, admin bool) types.Session {
	session := &SessionCtx{
		id:        id,
		admin:     admin,
		manager:   manager,
		logger:    manager.logger.With().Str("id", id).Logger(),
		connected: false,
	}

	manager.members[id] = session
	return session
}

func (manager *SessionManagerCtx) Get(id string) (types.Session, bool) {
	session, ok := manager.members[id]
	return session, ok
}

func (manager *SessionManagerCtx) Has(id string) bool {
	_, ok := manager.members[id]
	return ok
}

func (manager *SessionManagerCtx) Destroy(id string) error {
	session, ok := manager.members[id]
	if ok {
		delete(manager.members, id)
		err := session.destroy()

		manager.emmiter.Emit("destroy", id)
		return err
	}

	return nil
}

// ---
// host
// ---
func (manager *SessionManagerCtx) HasHost() bool {
	return manager.host != nil
}

func (manager *SessionManagerCtx) SetHost(host types.Session) {
	manager.host = host
	manager.emmiter.Emit("host", host)
}

func (manager *SessionManagerCtx) GetHost() types.Session {
	return manager.host
}

func (manager *SessionManagerCtx) ClearHost() {
	host := manager.host
	manager.host = nil
	manager.emmiter.Emit("host_cleared", host)
}

// ---
// members list
// ---
func (manager *SessionManagerCtx) Admins() []types.Session {
	var sessions []types.Session
	for _, session := range manager.members {
		if !session.connected || !session.admin {
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions
}

func (manager *SessionManagerCtx) Members() []types.Session {
	var sessions []types.Session
	for _, session := range manager.members {
		if !session.connected {
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions
}

func (manager *SessionManagerCtx) Broadcast(v interface{}, exclude interface{}) error {
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

func (manager *SessionManagerCtx) OnDestroy(listener func(id string)) {
	manager.emmiter.On("destroy", func(payload ...interface{}) {
		// Stop streaming, if everyone left
		if manager.capture.Streaming() && len(manager.members) == 0 {
			manager.capture.StopStream()
		}

		listener(payload[0].(string))
	})
}

func (manager *SessionManagerCtx) OnCreated(listener func(session types.Session)) {
	manager.emmiter.On("created", func(payload ...interface{}) {
		// Start streaming, when first joins
		if !manager.capture.Streaming() {
			manager.capture.StartStream()
		}
	
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnConnected(listener func(session types.Session)) {
	manager.emmiter.On("connected", func(payload ...interface{}) {
		listener(payload[0].(*SessionCtx))
	})
}
