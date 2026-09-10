package chat

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	appchat "github.com/m1k1o/neko/server/internal/application/chat"
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
		logger:   logger,
		config:   config,
		sessions: sessions,
		service:  appchat.NewService(sessions, config.Enabled, config.HistoryFile, config.HistoryLimit),
	}
}

type Manager struct {
	logger   zerolog.Logger
	config   *Config
	sessions types.SessionManager
	service  *appchat.Service
}

type Settings struct {
	CanSend    bool `json:"can_send" mapstructure:"can_send"`
	CanReceive bool `json:"can_receive" mapstructure:"can_receive"`
}

func (m *Manager) settingsForSession(session types.Session) (Settings, error) {
	settings, err := m.service.Settings(session)
	return Settings{CanSend: settings.CanSend, CanReceive: settings.CanReceive}, err
}

func (m *Manager) sendMessage(session types.Session, content Content) {
	_ = m.service.Send(session, appchat.Content(content))
}

func (m *Manager) Start() error {
	// send init message once a user connects
	m.sessions.OnConnected(func(session types.Session) {
		m.service.Initialize(session)
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
