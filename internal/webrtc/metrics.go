package webrtc

import (
	"sync"

	"github.com/pion/webrtc/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.com/demodesk/neko/server/pkg/types"
)

type metrics struct {
	connectionState      prometheus.Gauge
	connectionStateCount prometheus.Counter
	connectionCount      prometheus.Counter

	iceCandidates      map[string]struct{}
	iceCandidatesCount prometheus.Counter

	bytesSent     prometheus.Gauge
	bytesReceived prometheus.Gauge
}

type metricsCtx struct {
	mu sync.Mutex

	sessions map[string]metrics
}

func newMetrics() *metricsCtx {
	return &metricsCtx{
		sessions: map[string]metrics{},
	}
}

func (m *metricsCtx) getBySession(session types.Session) metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	met, ok := m.sessions[session.ID()]
	if ok {
		return met
	}

	met = metrics{
		connectionState: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "connection_state",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Connection state of session.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
			},
		}),
		connectionStateCount: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "connection_state_count",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Count of connection state changes for a session.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
			},
		}),
		connectionCount: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "connection_count",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Connection count of a session.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
			},
		}),

		iceCandidates: map[string]struct{}{},
		iceCandidatesCount: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "ice_candidates_count",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Count of ICE candidates sent by a remote client.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
			},
		}),

		bytesSent: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "bytes_sent",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Sent bytes to a session.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
			},
		}),
		bytesReceived: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "bytes_received",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Received bytes from a session.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
			},
		}),
	}

	m.sessions[session.ID()] = met
	return met
}

func (m *metricsCtx) NewConnection(session types.Session) {
	met := m.getBySession(session)
	met.connectionCount.Add(1)
}

func (m *metricsCtx) NewICECandidate(session types.Session, id string) {
	met := m.getBySession(session)

	if _, found := met.iceCandidates[id]; found {
		return
	}

	met.iceCandidates[id] = struct{}{}
	met.iceCandidatesCount.Add(1)
}

func (m *metricsCtx) SetState(session types.Session, state webrtc.PeerConnectionState) {
	met := m.getBySession(session)

	switch state {
	case webrtc.PeerConnectionStateNew:
		met.connectionState.Set(0)
	case webrtc.PeerConnectionStateConnecting:
		met.connectionState.Set(4)
	case webrtc.PeerConnectionStateConnected:
		met.connectionState.Set(5)
	case webrtc.PeerConnectionStateDisconnected:
		met.connectionState.Set(3)
	case webrtc.PeerConnectionStateFailed:
		met.connectionState.Set(2)
	case webrtc.PeerConnectionStateClosed:
		met.connectionState.Set(1)
	default:
		met.connectionState.Set(-1)
	}

	met.connectionStateCount.Add(1)
}

func (m *metricsCtx) SetTransportStats(session types.Session, data webrtc.TransportStats) {
	met := m.getBySession(session)

	met.bytesSent.Set(float64(data.BytesSent))
	met.bytesReceived.Set(float64(data.BytesReceived))
}
