package openinapp

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

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
	logger := log.With().Str("module", "openinapp").Logger()

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

func (m *Manager) Route(r types.Router) {
	r.Post("/openlink", m.openLinkHandler)
}

func (m *Manager) openLinkHandler(w http.ResponseWriter, r *http.Request) error {
	session, ok := auth.GetSession(r)
	if !ok {
		return utils.HttpUnauthorized("session not found")
	}

	if !m.config.Enabled {
		return utils.HttpForbidden("openinapp is disabled")
	}

	if !session.IsHost() {
		return utils.HttpForbidden("only the host can open links")
	}

	var req Url
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return utils.HttpBadRequest().WithInternalErr(err).Msg("failed to parse request body")
	}

	parsedUrl, err := url.Parse(req.Text)
	if err != nil {
		return utils.HttpBadRequest().WithInternalErr(err).Msg("failed to parse URL")
	}

	if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" {
		return utils.HttpBadRequest().Msg("URL must use http or https scheme")
	}

	parts := strings.Fields(m.config.OpenCommand)
	if len(parts) == 0 {
		return utils.HttpInternalServerError().Msg("open command is not configured")
	}

	if err := exec.Command(parts[0], append(parts[1:], parsedUrl.String())...).Start(); err != nil {
		return utils.HttpInternalServerError().WithInternalErr(err).Msg("failed to start open command")
	}

	return utils.HttpSuccess(w, nil)
}

func (m *Manager) WebSocketHandler(session types.Session, msg types.WebSocketMessage) bool {
	switch msg.Event {
	case OPENINAPP_OPENLINK:
		if !m.config.Enabled {
			m.logger.Warn().Msg("openlink request recieved while plugin is disabled")
			// we processed the openlink request, return true
			return true
		}

		var req Url
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			m.logger.Error().Err(err).Msg("failed to unmarshal openlink request")
			// we processed the openlink request, return true
			return true
		}

		if !session.IsHost() {
			m.logger.Warn().Msg("openlink request by non-host")
			// we processed the openlink request, return true
			return true
		}

		parsedUrl, err := url.Parse(req.Text)
		if err != nil {
			m.logger.Error().Err(err).Msg("failed to parse openlink request")
			// we processed the openlink request, return true
			return true
		}

		if !(parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https") {
			m.logger.Warn().Msg("openlink url does not have valid scheme")
			// we processed the openlink url, return true
			return true
		}

		parts := strings.Fields(m.config.OpenCommand)
		if len(parts) == 0 {
			m.logger.Error().Msg("open command is not configured")
			return true
		}
		if err := exec.Command(parts[0], append(parts[1:], parsedUrl.String())...).Start(); err != nil {
			m.logger.Error().Err(err).Msg("failed to start open command")
			// we processed the openlink request, return true
			return true
		}

		return true
	}
	return false
}

func (m *Manager) Start() error {
	// send init message once a user connects
	m.sessions.OnConnected(func(session types.Session) {
		session.Send(OPENINAPP_INIT, Init{
			Enabled: m.config.Enabled,
		})
	})

	return nil
}

func (m *Manager) Shutdown() error {
	return nil
}
