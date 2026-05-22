package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Tunables; small enough that the worker can't block neko's session
// goroutines, large enough that a brief api downtime doesn't lose events.
const (
	queueDepth         = 256
	defaultHTTPTimeout = 5 * time.Second
)

// Manager subscribes to session connect/disconnect events and forwards them
// to kernel-images-api as live_view_* telemetry events. All HTTP work runs
// on a single background worker so neko's emitter callbacks stay non-blocking.
type Manager struct {
	logger     zerolog.Logger
	config     *Config
	sessions   types.SessionManager
	httpClient *http.Client

	mu          sync.Mutex
	connectedAt map[string]time.Time

	eventsCh chan eventPayload
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewManager builds a manager with sensible defaults; tests can override
// httpClient via the exported field after construction.
func NewManager(sessions types.SessionManager, config *Config) *Manager {
	return &Manager{
		logger:      log.With().Str("module", "telemetry").Logger(),
		config:      config,
		sessions:    sessions,
		httpClient:  &http.Client{Timeout: defaultHTTPTimeout},
		connectedAt: make(map[string]time.Time),
		eventsCh:    make(chan eventPayload, queueDepth),
		stopCh:      make(chan struct{}),
	}
}

// Start subscribes to session events and launches the publish worker. No-op
// when the plugin is disabled.
func (m *Manager) Start() error {
	if !m.config.Enabled {
		return nil
	}
	m.logger.Info().Str("endpoint", m.config.Endpoint).Msg("plugin enabled")

	m.wg.Add(1)
	go m.worker()

	m.sessions.OnConnected(func(session types.Session) {
		m.handleConnect(session.ID())
	})
	m.sessions.OnDisconnected(func(session types.Session) {
		m.handleDisconnect(session.ID())
	})

	return nil
}

// Shutdown signals the worker to drain and exit, blocking until it does.
func (m *Manager) Shutdown() error {
	if !m.config.Enabled {
		return nil
	}
	close(m.stopCh)
	m.wg.Wait()
	return nil
}

func (m *Manager) handleConnect(id string) {
	m.mu.Lock()
	m.connectedAt[id] = time.Now()
	m.mu.Unlock()

	m.enqueue(eventPayload{
		Type: "live_view_connect",
		Data: map[string]any{"session_id": id},
	})
}

func (m *Manager) handleDisconnect(id string) {
	m.mu.Lock()
	start, ok := m.connectedAt[id]
	delete(m.connectedAt, id)
	m.mu.Unlock()

	var durationMs float64
	if ok {
		durationMs = float64(time.Since(start).Microseconds()) / 1000.0
	}

	m.enqueue(eventPayload{
		Type: "live_view_disconnect",
		Data: map[string]any{"session_id": id, "duration_ms": durationMs},
	})
}

func (m *Manager) enqueue(ev eventPayload) {
	select {
	case m.eventsCh <- ev:
	default:
		// Drop rather than block neko's session goroutines. A backed-up
		// kernel-images-api means we'd lose lifecycle pairs anyway.
		m.logger.Warn().Str("type", ev.Type).Msg("telemetry queue full; dropping event")
	}
}

func (m *Manager) worker() {
	defer m.wg.Done()
	for {
		select {
		case <-m.stopCh:
			// Best-effort drain of remaining events on shutdown so we don't
			// lose paired connect/disconnects when neko exits cleanly.
			for {
				select {
				case ev := <-m.eventsCh:
					m.publish(ev)
				default:
					return
				}
			}
		case ev := <-m.eventsCh:
			m.publish(ev)
		}
	}
}

func (m *Manager) publish(ev eventPayload) {
	body := publishBody{
		Type:     ev.Type,
		Category: "system",
		Source: publishSource{
			Kind:  "local_process",
			Event: "neko." + ev.Type,
		},
		Data: ev.Data,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		m.logger.Warn().Err(err).Msg("marshal telemetry body failed")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultHTTPTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.config.Endpoint, bytes.NewReader(raw))
	if err != nil {
		m.logger.Warn().Err(err).Msg("build telemetry request failed")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.httpClient.Do(req)
	if err != nil {
		m.logger.Debug().Err(err).Str("type", ev.Type).Msg("telemetry POST failed")
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		m.logger.Debug().Int("status", resp.StatusCode).Str("type", ev.Type).Msg("telemetry POST non-2xx")
	}
}

type eventPayload struct {
	Type string
	Data map[string]any
}

type publishSource struct {
	Kind  string `json:"kind"`
	Event string `json:"event,omitempty"`
}

type publishBody struct {
	Type     string         `json:"type"`
	Category string         `json:"category"`
	Source   publishSource  `json:"source"`
	Data     map[string]any `json:"data"`
}
