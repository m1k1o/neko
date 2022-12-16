package webrtc

import (
	"sync"

	"github.com/demodesk/neko/pkg/types"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	connectionState      prometheus.Gauge
	connectionStateCount prometheus.Counter
	connectionCount      prometheus.Counter

	iceCandidates         map[string]struct{}
	iceCandidatesMu       *sync.Mutex
	iceCandidatesUdpCount prometheus.Counter
	iceCandidatesTcpCount prometheus.Counter

	iceCandidatesUsedUdp prometheus.Gauge
	iceCandidatesUsedTcp prometheus.Gauge

	videoIds   map[string]prometheus.Gauge
	videoIdsMu *sync.Mutex

	receiverEstimatedMaximumBitrate prometheus.Gauge

	receiverReportDelay     prometheus.Gauge
	receiverReportJitter    prometheus.Gauge
	receiverReportTotalLost prometheus.Gauge

	iceBytesSent      prometheus.Gauge
	iceBytesReceived  prometheus.Gauge
	sctpBytesSent     prometheus.Gauge
	sctpBytesReceived prometheus.Gauge
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

		iceCandidates:   map[string]struct{}{},
		iceCandidatesMu: &sync.Mutex{},
		iceCandidatesUdpCount: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "ice_candidates_count",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Count of ICE candidates sent by a remote client.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
				"protocol":   "udp",
			},
		}),
		iceCandidatesTcpCount: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "ice_candidates_count",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Count of ICE candidates sent by a remote client.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
				"protocol":   "tcp",
			},
		}),

		iceCandidatesUsedUdp: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "ice_candidates_used",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Used ICE candidates that are currently in use.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
				"protocol":   "udp",
			},
		}),
		iceCandidatesUsedTcp: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "ice_candidates_used",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Used ICE candidates that are currently in use.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
				"protocol":   "tcp",
			},
		}),

		videoIds:   map[string]prometheus.Gauge{},
		videoIdsMu: &sync.Mutex{},

		receiverEstimatedMaximumBitrate: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "receiver_estimated_maximum_bitrate",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Receiver Estimated Maximum Bitrate from SCTP.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
			},
		}),

		receiverReportDelay: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "receiver_report_delay",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Receiver Report Delay from SCTP, expressed in units of 1/65536 seconds.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
			},
		}),
		receiverReportJitter: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "receiver_report_jitter",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Receiver Report Jitter from SCTP.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
			},
		}),
		receiverReportTotalLost: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "receiver_report_total_lost",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Receiver Report Total Lost from SCTP.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
			},
		}),

		iceBytesSent: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "bytes_sent",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Sent bytes to a session.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
				"transport":  "ice",
			},
		}),
		iceBytesReceived: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "bytes_received",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Received bytes from a session.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
				"transport":  "ice",
			},
		}),

		sctpBytesSent: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "bytes_sent",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Sent bytes to a session.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
				"transport":  "sctp",
			},
		}),
		sctpBytesReceived: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "bytes_received",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Received bytes from a session.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
				"transport":  "sctp",
			},
		}),
	}

	m.sessions[session.ID()] = met
	return met
}

func (m *metricsCtx) reset(met metrics) {
	met.videoIdsMu.Lock()
	for _, entry := range met.videoIds {
		entry.Set(0)
	}
	met.videoIdsMu.Unlock()

	met.iceCandidatesUsedUdp.Set(float64(0))
	met.iceCandidatesUsedTcp.Set(float64(0))

	met.receiverEstimatedMaximumBitrate.Set(0)

	met.receiverReportDelay.Set(0)
	met.receiverReportJitter.Set(0)
}

func (m *metricsCtx) NewConnection(session types.Session) {
	met := m.getBySession(session)
	met.connectionCount.Add(1)
}

func (m *metricsCtx) NewICECandidate(session types.Session, candidate webrtc.ICECandidateStats) {
	met := m.getBySession(session)

	met.iceCandidatesMu.Lock()
	defer met.iceCandidatesMu.Unlock()

	if _, found := met.iceCandidates[candidate.ID]; found {
		return
	}

	met.iceCandidates[candidate.ID] = struct{}{}
	if candidate.Protocol == "udp" {
		met.iceCandidatesUdpCount.Add(1)
	} else if candidate.Protocol == "tcp" {
		met.iceCandidatesTcpCount.Add(1)
	}
}

func (m *metricsCtx) SetICECandidatesUsed(session types.Session, candidates []webrtc.ICECandidateStats) {
	met := m.getBySession(session)

	udp, tcp := 0, 0
	for _, candidate := range candidates {
		if candidate.Protocol == "udp" {
			udp++
		} else if candidate.Protocol == "tcp" {
			tcp++
		}
	}

	met.iceCandidatesUsedUdp.Set(float64(udp))
	met.iceCandidatesUsedTcp.Set(float64(tcp))
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
		m.reset(met)
	default:
		met.connectionState.Set(-1)
	}

	met.connectionStateCount.Add(1)
}

func (m *metricsCtx) SetVideoID(session types.Session, videoId string) {
	met := m.getBySession(session)

	met.videoIdsMu.Lock()
	defer met.videoIdsMu.Unlock()

	if _, found := met.videoIds[videoId]; !found {
		met.videoIds[videoId] = promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "video_listeners",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Listeners for Video pipelines by a session.",
			ConstLabels: map[string]string{
				"session_id": session.ID(),
				"video_id":   videoId,
			},
		})
	}

	for id, entry := range met.videoIds {
		if id == videoId {
			entry.Set(1)
		} else {
			entry.Set(0)
		}
	}
}

func (m *metricsCtx) SetReceiverEstimatedMaximumBitrate(session types.Session, bitrate float32) {
	met := m.getBySession(session)

	met.receiverEstimatedMaximumBitrate.Set(float64(bitrate))
}

func (m *metricsCtx) SetReceiverReport(session types.Session, report rtcp.ReceptionReport) {
	met := m.getBySession(session)

	met.receiverReportDelay.Set(float64(report.Delay))
	met.receiverReportJitter.Set(float64(report.Jitter))
	met.receiverReportTotalLost.Set(float64(report.TotalLost))
}

func (m *metricsCtx) SetIceTransportStats(session types.Session, data webrtc.TransportStats) {
	met := m.getBySession(session)

	met.iceBytesSent.Set(float64(data.BytesSent))
	met.iceBytesReceived.Set(float64(data.BytesReceived))
}

func (m *metricsCtx) SetSctpTransportStats(session types.Session, data webrtc.TransportStats) {
	met := m.getBySession(session)

	met.sctpBytesSent.Set(float64(data.BytesSent))
	met.sctpBytesReceived.Set(float64(data.BytesReceived))
}
