package scaletozero

import (
	"context"
	"sync"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/onkernel/kernel-images/server/lib/scaletozero"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func NewManager(
	sessions types.SessionManager,
	config *Config,
) *Manager {
	logger := log.With().Str("module", "scaletozero").Logger()

	return &Manager{
		logger:   logger,
		config:   config,
		sessions: sessions,
		ctrl:     scaletozero.NewUnikraftCloudController(),
	}
}

type Manager struct {
	logger   zerolog.Logger
	config   *Config
	sessions types.SessionManager
	ctrl     scaletozero.Controller
	mu       sync.Mutex
	shutdown bool
	pending  int
}

func (m *Manager) Start() error {
	if !m.config.Enabled {
		return nil
	}
	m.logger.Info().Msg("scale-to-zero plugin enabled")

	m.sessions.OnConnected(func(session types.Session) {
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.shutdown {
			return
		}

		m.pending++
		m.ctrl.Disable(context.Background())
	})

	m.sessions.OnDisconnected(func(session types.Session) {
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.shutdown {
			return
		}
		m.pending--
		m.ctrl.Enable(context.Background())
	})

	return nil
}

func (m *Manager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdown = true

	for i := 0; i < m.pending; i++ {
		m.ctrl.Enable(context.Background())
	}

	return nil
}
