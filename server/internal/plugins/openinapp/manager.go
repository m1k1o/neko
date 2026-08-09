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

	if !session.IsHost() && (!session.Profile().CanHost || !m.sessions.Settings().ImplicitHosting) {
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
