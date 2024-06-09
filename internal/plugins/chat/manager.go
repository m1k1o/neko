package chat

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/pkg/auth"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

func NewManager(
	sessions types.SessionManager,
	config *Config,
) *Manager {
	logger := log.With().Str("module", "chat").Logger()

	return &Manager{
		logger:   logger,
		config:   config,
		sessions: sessions,
	}
}

type Manager struct {
	logger   zerolog.Logger
	config   *Config
	sessions types.SessionManager
}

func (m *Manager) isEnabledForSession(session types.Session) bool {
	return m.config.Enabled &&
		settingsIsEnabled(m.sessions.Settings()) &&
		profileIsEnabled(session.Profile())
}

func (m *Manager) sendMessage(session types.Session, content Content) {
	now := time.Now()

	// get all sessions that have chat enabled
	var sessions []types.Session
	m.sessions.Range(func(s types.Session) bool {
		if m.isEnabledForSession(s) {
			sessions = append(sessions, s)
		}
		// continue iteration over all sessions
		return true
	})

	// send content to all sessions
	for _, s := range sessions {
		s.Send(CHAT_MESSAGE, Message{
			ID:      session.ID(),
			Created: now,
			Content: content,
		})
	}
}

func (m *Manager) Start() error {
	// send init message once a user connects
	m.sessions.OnConnected(func(session types.Session) {
		isEnabled := m.isEnabledForSession(session)

		// send init message
		session.Send(CHAT_INIT, Init{
			Enabled: isEnabled,
		})
	})

	// do not proceed if chat is disabled in the config
	if !m.config.Enabled {
		return nil
	}

	// on settings change, reinit if chat is enabled/disabled
	m.sessions.OnSettingsChanged(func(session types.Session, new, old types.Settings) {
		isEnabled := settingsIsEnabled(new)
		wasEnabled := settingsIsEnabled(old)

		if !isEnabled && wasEnabled {
			// if chat was enabled and is now disabled, broadcast to all sessions
			// because it cannot be overridden by profile settings
			m.sessions.Broadcast(CHAT_INIT, Init{
				Enabled: false,
			})
		}
		if isEnabled && !wasEnabled {
			// if chat was disabled and is now enabled, loop over all sessions
			// and send the init message (because it can be overridden by profile settings)
			for _, s := range m.sessions.List() {
				s.Send(CHAT_INIT, Init{
					Enabled: m.isEnabledForSession(s),
				})
			}
		}
	})

	// on profile change, reinit if chat is enabled/disabled
	m.sessions.OnProfileChanged(func(session types.Session, new, old types.MemberProfile) {
		isEnabled := profileIsEnabled(new)
		wasEnabled := profileIsEnabled(old)

		if isEnabled != wasEnabled {
			// only if the chat setting was changed, send the init message
			session.Send(CHAT_INIT, Init{
				Enabled: m.isEnabledForSession(session),
			})
		}
	})

	return nil
}

func (m *Manager) Shutdown() error {
	return nil
}

func (m *Manager) Route(r types.Router) {
	r.With(auth.AdminsOnly).Post("/", m.sendMessageHandler)
}

func (m *Manager) WebSocketHandler(session types.Session, msg types.WebSocketMessage) bool {
	switch msg.Event {
	case CHAT_MESSAGE:
		var content Content
		if err := json.Unmarshal(msg.Payload, &content); err != nil {
			m.logger.Error().Err(err).Msg("failed to unmarshal chat message")
			// we processed the message, return true
			return true
		}

		m.sendMessage(session, content)
		return true
	}
	return false
}

func (m *Manager) sendMessageHandler(w http.ResponseWriter, r *http.Request) error {
	session, ok := auth.GetSession(r)
	if !ok {
		return utils.HttpUnauthorized("session not found")
	}

	enabled := m.isEnabledForSession(session)
	if !enabled {
		return utils.HttpForbidden("chat is disabled")
	}

	content := Content{}
	if err := utils.HttpJsonRequest(w, r, &content); err != nil {
		return err
	}

	m.sendMessage(session, content)
	return utils.HttpSuccess(w)
}

func settingsIsEnabled(s types.Settings) bool {
	isEnabled := true

	settings, ok := s.Plugins["chat"]
	// by default, allow chat if the plugin config is not present
	if ok {
		isEnabled, ok = settings.(bool)
		// if the plugin is present but not a boolean, allow chat
		if !ok {
			isEnabled = true
		}
	}

	return isEnabled
}

func profileIsEnabled(p types.MemberProfile) bool {
	isEnabled := true

	settings, ok := p.Plugins["chat"]
	// by default, allow chat if the plugin config is not present
	if ok {
		isEnabled, ok = settings.(bool)
		// if the plugin is present but not a boolean, allow chat
		if !ok {
			isEnabled = true
		}
	}

	return isEnabled
}
