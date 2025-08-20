package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/auth"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

func NewManager(
	sessions types.SessionManager,
	config *Config,
) *Manager {
	logger := log.With().Str("module", "chat").Logger()

	return &Manager{
		logger:         logger,
		config:         config,
		sessions:       sessions,
		messageHistory: make([]Message, 0),
	}
}

type Manager struct {
	logger         zerolog.Logger
	config         *Config
	sessions       types.SessionManager
	messageHistory []Message
}

type Settings struct {
	CanSend    bool `json:"can_send" mapstructure:"can_send"`
	CanReceive bool `json:"can_receive" mapstructure:"can_receive"`
}

func (m *Manager) settingsForSession(session types.Session) (Settings, error) {
	settings := Settings{
		CanSend:    true, // defaults to true
		CanReceive: true, // defaults to true
	}
	err := m.sessions.Settings().Plugins.Unmarshal(PluginName, &settings)
	if err != nil && !errors.Is(err, types.ErrPluginSettingsNotFound) {
		return Settings{}, fmt.Errorf("unable to unmarshal %s plugin settings from global settings: %w", PluginName, err)
	}

	profile := Settings{
		CanSend:    true, // defaults to true
		CanReceive: true, // defaults to true
	}

	err = session.Profile().Plugins.Unmarshal(PluginName, &profile)
	if err != nil && !errors.Is(err, types.ErrPluginSettingsNotFound) {
		return Settings{}, fmt.Errorf("unable to unmarshal %s plugin settings from profile: %w", PluginName, err)
	}

	return Settings{
		CanSend:    m.config.Enabled && (settings.CanSend || session.Profile().IsAdmin) && profile.CanSend,
		CanReceive: m.config.Enabled && (settings.CanReceive || session.Profile().IsAdmin) && profile.CanReceive,
	}, nil
}

func (m *Manager) sendMessage(session types.Session, content Content) {
	// Create the message
	message := Message{
		ID:      session.ID(),
		Created: time.Now(),
		Content: content,
	}
	m.addMessageToHistory(message)

	// get all sessions that have chat enabled
	var sessions []types.Session
	m.sessions.Range(func(s types.Session) bool {
		if settings, err := m.settingsForSession(s); err == nil && settings.CanReceive {
			sessions = append(sessions, s)
		}
		// continue iteration over all sessions
		return true
	})

	// send content to all sessions
	for _, s := range sessions {
		s.Send(CHAT_MESSAGE, message)
	}
}

func (m *Manager) addMessageToHistory(message Message) {
	// Add message to history
	m.messageHistory = append(m.messageHistory, message)

	// Keep only the last 100 messages
	if len(m.messageHistory) > 100 {
		m.messageHistory = m.messageHistory[len(m.messageHistory)-100:]
	}
}

func (m *Manager) Start() error {
	// send init message once a user connects
	m.sessions.OnConnected(func(session types.Session) {
		messageHistory := m.messageHistory
		// Check if user can receive messages and reset the message history if he can't
		settings, err := m.settingsForSession(session)
		if err != nil || !settings.CanReceive {
			messageHistory = make([]Message, 0)
		}

		// Send init message with message history
		session.Send(CHAT_INIT, Init{
			Enabled: m.config.Enabled,
			History: messageHistory,
		})
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

		settings, err := m.settingsForSession(session)
		if err != nil {
			m.logger.Error().Err(err).Msg("error checking chat permissions for this session")
			// we processed the message, return true
			return true
		}
		if !settings.CanSend {
			m.logger.Warn().Msg("not allowed to send chat messages")
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

	settings, err := m.settingsForSession(session)
	if err != nil {
		return utils.HttpInternalServerError().
			WithInternalErr(err).
			Msg("error checking chat permissions for this session")
	}

	if !settings.CanSend {
		return utils.HttpForbidden("not allowed to send chat messages")
	}

	content := Content{}
	if err := utils.HttpJsonRequest(w, r, &content); err != nil {
		return err
	}

	m.sendMessage(session, content)
	return utils.HttpSuccess(w)
}
