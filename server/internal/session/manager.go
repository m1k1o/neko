package session

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/utils"
)

func New(capture types.CaptureManager) *SessionManager {
	return &SessionManager{
		logger:        log.With().Str("module", "session").Logger(),
		host:          "",
		capture:       capture,
		eventsChannel: make(chan types.SessionEvent, 10),
		members:       make(map[string]*Session),
	}
}

type SessionManager struct {
	mu            sync.Mutex
	logger        zerolog.Logger
	host          string
	capture       types.CaptureManager
	members       map[string]*Session
	eventsChannel chan types.SessionEvent
	// TODO: Handle locks in sessions as flags.
	controlLocked bool
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
	manager.capture.Audio().AddListener()
	manager.capture.Video().AddListener()
	manager.mu.Unlock()

	manager.eventsChannel <- types.SessionEvent{
		Type:    types.SESSION_CREATED,
		Id:      id,
		Session: session,
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
	manager.mu.Lock()
	_, ok := manager.members[id]
	manager.mu.Unlock()

	if ok {
		manager.host = id

		manager.eventsChannel <- types.SessionEvent{
			Type: types.SESSION_HOST_SET,
			Id:   id,
		}

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

	manager.eventsChannel <- types.SessionEvent{
		Type: types.SESSION_HOST_CLEARED,
		Id:   id,
	}
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

// TODO: Handle locks in sessions as flags.
func (manager *SessionManager) SetControlLocked(locked bool) {
	manager.controlLocked = locked
}

func (manager *SessionManager) CanControl(id string) bool {
	session, ok := manager.Get(id)
	return ok && (!manager.controlLocked || session.Admin())
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

		manager.capture.Audio().RemoveListener()
		manager.capture.Video().RemoveListener()
		manager.mu.Unlock()

		manager.eventsChannel <- types.SessionEvent{
			Type:    types.SESSION_DESTROYED,
			Id:      id,
			Session: session,
		}
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

func (manager *SessionManager) AdminBroadcast(v interface{}, exclude interface{}) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for id, session := range manager.members {
		if !session.connected || !session.admin {
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

func (manager *SessionManager) GetEventsChannel() chan types.SessionEvent {
	return manager.eventsChannel
}
