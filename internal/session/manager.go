package session

import (
	"errors"
	"sync"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/internal/config"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

func New(config *config.Session) *SessionManagerCtx {
	manager := &SessionManagerCtx{
		logger: log.With().Str("module", "session").Logger(),
		config: config,
		settings: types.Settings{
			PrivateMode:       false, // By default disabled.
			ImplicitHosting:   config.ImplicitHosting,
			InactiveCursors:   config.InactiveCursors,
			MercifulReconnect: config.MercifulReconnect,
		},
		tokens:   make(map[string]string),
		sessions: make(map[string]*SessionCtx),
		cursors:  make(map[types.Session][]types.Cursor),
		emmiter:  events.New(),
	}

	// create API session
	if config.APIToken != "" {
		manager.apiSession = &SessionCtx{
			id:      "API",
			token:   config.APIToken,
			manager: manager,
			logger:  manager.logger.With().Str("session_id", "API").Logger(),
			profile: types.MemberProfile{
				Name:               "API Session",
				IsAdmin:            true,
				CanLogin:           true,
				CanConnect:         false,
				CanWatch:           true,
				CanHost:            true,
				CanAccessClipboard: true,
			},
		}
	}

	return manager
}

type SessionManagerCtx struct {
	logger zerolog.Logger
	config *config.Session

	settings   types.Settings
	settingsMu sync.Mutex

	tokens     map[string]string
	sessions   map[string]*SessionCtx
	sessionsMu sync.Mutex

	host   types.Session
	hostMu sync.Mutex

	cursors   map[types.Session][]types.Cursor
	cursorsMu sync.Mutex

	emmiter    events.EventEmmiter
	apiSession *SessionCtx
}

func (manager *SessionManagerCtx) Create(id string, profile types.MemberProfile) (types.Session, string, error) {
	token, err := utils.NewUID(64)
	if err != nil {
		return nil, "", err
	}

	manager.sessionsMu.Lock()
	if _, ok := manager.sessions[id]; ok {
		manager.sessionsMu.Unlock()
		return nil, "", types.ErrSessionAlreadyExists
	}

	if _, ok := manager.tokens[token]; ok {
		manager.sessionsMu.Unlock()
		return nil, "", errors.New("session token already exists")
	}

	session := &SessionCtx{
		id:      id,
		token:   token,
		manager: manager,
		logger:  manager.logger.With().Str("session_id", id).Logger(),
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
		return types.ErrSessionNotFound
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
		return types.ErrSessionNotFound
	}

	delete(manager.tokens, session.token)
	delete(manager.sessions, id)
	manager.sessionsMu.Unlock()

	if session.State().IsConnected {
		session.GetWebSocketPeer().Destroy("session deleted")
	}

	if session.State().IsWatching {
		session.GetWebRTCPeer().Destroy()
	}

	manager.emmiter.Emit("deleted", session)
	return nil
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

	if ok {
		return manager.Get(id)
	}

	// is API session
	if manager.apiSession != nil && manager.apiSession.token == token {
		return manager.apiSession, true
	}

	return nil, false
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
// cursors
// ---

func (manager *SessionManagerCtx) SetCursor(cursor types.Cursor, session types.Session) {
	manager.cursorsMu.Lock()
	defer manager.cursorsMu.Unlock()

	list, ok := manager.cursors[session]
	if !ok {
		list = []types.Cursor{}
	}

	list = append(list, cursor)
	manager.cursors[session] = list
}

func (manager *SessionManagerCtx) PopCursors() map[types.Session][]types.Cursor {
	manager.cursorsMu.Lock()
	defer manager.cursorsMu.Unlock()

	cursors := manager.cursors
	manager.cursors = make(map[types.Session][]types.Cursor)

	return cursors
}

// ---
// broadcasts
// ---

func (manager *SessionManagerCtx) Broadcast(event string, payload any, exclude ...string) {
	for _, session := range manager.List() {
		if !session.State().IsConnected {
			continue
		}

		if len(exclude) > 0 {
			if in, _ := utils.ArrayIn(session.ID(), exclude); in {
				continue
			}
		}

		session.Send(event, payload)
	}
}

func (manager *SessionManagerCtx) AdminBroadcast(event string, payload any, exclude ...string) {
	for _, session := range manager.List() {
		if !session.State().IsConnected || !session.Profile().IsAdmin {
			continue
		}

		if len(exclude) > 0 {
			if in, _ := utils.ArrayIn(session.ID(), exclude); in {
				continue
			}
		}

		session.Send(event, payload)
	}
}

func (manager *SessionManagerCtx) InactiveCursorsBroadcast(event string, payload any, exclude ...string) {
	for _, session := range manager.List() {
		if !session.State().IsConnected || !session.Profile().CanSeeInactiveCursors {
			continue
		}

		if len(exclude) > 0 {
			if in, _ := utils.ArrayIn(session.ID(), exclude); in {
				continue
			}
		}

		session.Send(event, payload)
	}
}

// ---
// events
// ---

func (manager *SessionManagerCtx) OnCreated(listener func(session types.Session)) {
	manager.emmiter.On("created", func(payload ...any) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnDeleted(listener func(session types.Session)) {
	manager.emmiter.On("deleted", func(payload ...any) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnConnected(listener func(session types.Session)) {
	manager.emmiter.On("connected", func(payload ...any) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnDisconnected(listener func(session types.Session)) {
	manager.emmiter.On("disconnected", func(payload ...any) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnProfileChanged(listener func(session types.Session)) {
	manager.emmiter.On("profile_changed", func(payload ...any) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnStateChanged(listener func(session types.Session)) {
	manager.emmiter.On("state_changed", func(payload ...any) {
		listener(payload[0].(*SessionCtx))
	})
}

func (manager *SessionManagerCtx) OnHostChanged(listener func(session types.Session)) {
	manager.emmiter.On("host_changed", func(payload ...any) {
		if payload[0] == nil {
			listener(nil)
		} else {
			listener(payload[0].(*SessionCtx))
		}
	})
}

func (manager *SessionManagerCtx) OnSettingsChanged(listener func(new types.Settings, old types.Settings)) {
	manager.emmiter.On("settings_changed", func(payload ...any) {
		listener(payload[0].(types.Settings), payload[1].(types.Settings))
	})
}

// ---
// settings
// ---

func (manager *SessionManagerCtx) UpdateSettings(new types.Settings) {
	manager.settingsMu.Lock()
	old := manager.settings
	manager.settings = new
	manager.settingsMu.Unlock()

	// if private mode changed
	if old.PrivateMode != new.PrivateMode {
		// update webrtc paused state for all sessions
		for _, session := range manager.List() {
			enabled := session.PrivateModeEnabled()

			// if session had control, it must release it
			if enabled && session.IsHost() {
				manager.ClearHost()
			}

			// its webrtc connection will be paused or unpaused
			if webrtcPeer := session.GetWebRTCPeer(); webrtcPeer != nil {
				webrtcPeer.SetPaused(enabled)
			}
		}
	}

	manager.emmiter.Emit("settings_changed", new, old)
}

func (manager *SessionManagerCtx) Settings() types.Settings {
	manager.settingsMu.Lock()
	defer manager.settingsMu.Unlock()

	return manager.settings
}

func (manager *SessionManagerCtx) CookieEnabled() bool {
	return manager.config.CookieEnabled
}
