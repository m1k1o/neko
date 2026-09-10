package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type HealthReason string

const (
	HealthAuthenticationFailed HealthReason = "authentication_failed"
	HealthUpstreamUnreachable  HealthReason = "upstream_unreachable"
	HealthTargetRejected       HealthReason = "target_rejected"
	HealthCheckFailed          HealthReason = "check_failed"
)

type CheckError struct {
	Reason HealthReason
	Err    error
}

func (e *CheckError) Error() string { return string(e.Reason) }
func (e *CheckError) Unwrap() error { return e.Err }

type HealthSnapshot struct {
	Status    string       `json:"status"`
	Protocol  Protocol     `json:"protocol"`
	Upstream  string       `json:"upstream"`
	Target    string       `json:"target"`
	CheckedAt time.Time    `json:"checked_at"`
	Reason    HealthReason `json:"reason,omitempty"`
}

// HealthMonitor actively verifies that the configured upstream proxy can
// authenticate and open a connection to the deployment-provided target.
type HealthMonitor struct {
	agent    *Agent
	endpoint Endpoint
	target   string

	mu       sync.RWMutex
	snapshot HealthSnapshot
}

func NewHealthMonitor(agent *Agent, endpoint Endpoint, target string) (*HealthMonitor, error) {
	if _, err := TargetAddress(target); err != nil {
		return nil, err
	}
	return &HealthMonitor{
		agent:    agent,
		endpoint: endpoint,
		target:   target,
		snapshot: HealthSnapshot{
			Status:   "unhealthy",
			Protocol: endpoint.Protocol,
			Upstream: endpoint.RedactedServer(),
			Target:   target,
			Reason:   HealthCheckFailed,
		},
	}, nil
}

func (m *HealthMonitor) Check(ctx context.Context) error {
	conn, err := m.agent.dialTunnel(ctx, m.target)
	checkedAt := time.Now().UTC()
	if err == nil {
		conn.Close()
		m.setSnapshot(HealthSnapshot{
			Status:    "healthy",
			Protocol:  m.endpoint.Protocol,
			Upstream:  m.endpoint.RedactedServer(),
			Target:    m.target,
			CheckedAt: checkedAt,
		})
		return nil
	}

	reason := classifyCheckError(err)
	m.setSnapshot(HealthSnapshot{
		Status:    "unhealthy",
		Protocol:  m.endpoint.Protocol,
		Upstream:  m.endpoint.RedactedServer(),
		Target:    m.target,
		CheckedAt: checkedAt,
		Reason:    reason,
	})
	return &CheckError{Reason: reason, Err: err}
}

func (m *HealthMonitor) Run(ctx context.Context, interval, timeout time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkContext, cancel := context.WithTimeout(ctx, timeout)
			_ = m.Check(checkContext)
			cancel()
		}
	}
}

func (m *HealthMonitor) Snapshot() HealthSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.snapshot
}

func (m *HealthMonitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/healthz" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snapshot := m.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	if snapshot.Status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(snapshot)
}

func (m *HealthMonitor) setSnapshot(snapshot HealthSnapshot) {
	m.mu.Lock()
	m.snapshot = snapshot
	m.mu.Unlock()
}

func classifyCheckError(err error) HealthReason {
	var checkError *CheckError
	if errors.As(err, &checkError) {
		return checkError.Reason
	}

	message := strings.ToLower(err.Error())
	if strings.Contains(message, "authentication failed") ||
		strings.Contains(message, "no acceptable authentication methods") ||
		strings.Contains(message, "unsupported authentication method") {
		return HealthAuthenticationFailed
	}
	if strings.Contains(message, "unknown error connection not allowed") ||
		strings.Contains(message, "unknown error network unreachable") ||
		strings.Contains(message, "unknown error host unreachable") ||
		strings.Contains(message, "unknown error connection refused") {
		return HealthTargetRejected
	}

	var operationError *net.OpError
	if errors.As(err, &operationError) {
		return HealthUpstreamUnreachable
	}
	return HealthCheckFailed
}
