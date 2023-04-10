package webrtc

import (
	"sync"

	"github.com/demodesk/neko/pkg/types"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metricsManager struct {
	mu sync.Mutex

	sessions map[string]metrics
}

func newMetricsManager() *metricsManager {
	return &metricsManager{
		sessions: map[string]metrics{},
	}
}

func (m *metricsManager) getBySession(session types.Session) metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionId := session.ID()

	met, ok := m.sessions[sessionId]
	if ok {
		return met
	}

	met = metrics{
		sessionId: sessionId,

		connectionState: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "connection_state",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Connection state of session.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
			},
		}),
		connectionStateCount: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "connection_state_count",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Count of connection state changes for a session.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
			},
		}),
		connectionCount: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "connection_count",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Connection count of a session.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
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
				"session_id": sessionId,
				"protocol":   "udp",
			},
		}),
		iceCandidatesTcpCount: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "ice_candidates_count",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Count of ICE candidates sent by a remote client.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
				"protocol":   "tcp",
			},
		}),

		iceCandidatesUsedUdp: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "ice_candidates_used",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Used ICE candidates that are currently in use.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
				"protocol":   "udp",
			},
		}),
		iceCandidatesUsedTcp: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "ice_candidates_used",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Used ICE candidates that are currently in use.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
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
				"session_id": sessionId,
			},
		}),
		receiverEstimatedTargetBitrate: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "receiver_estimated_target_bitrate",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Receiver Estimated Target Bitrate using Google's congestion control.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
			},
		}),

		receiverReportDelay: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "receiver_report_delay",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Receiver Report Delay from SCTP, expressed in units of 1/65536 seconds.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
			},
		}),
		receiverReportJitter: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "receiver_report_jitter",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Receiver Report Jitter from SCTP.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
			},
		}),
		receiverReportTotalLost: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "receiver_report_total_lost",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Receiver Report Total Lost from SCTP.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
			},
		}),

		iceBytesSent: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "bytes_sent",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Sent bytes to a session.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
				"transport":  "ice",
			},
		}),
		iceBytesReceived: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "bytes_received",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Received bytes from a session.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
				"transport":  "ice",
			},
		}),

		sctpBytesSent: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "bytes_sent",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Sent bytes to a session.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
				"transport":  "sctp",
			},
		}),
		sctpBytesReceived: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "bytes_received",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Received bytes from a session.",
			ConstLabels: map[string]string{
				"session_id": sessionId,
				"transport":  "sctp",
			},
		}),
	}

	m.sessions[sessionId] = met
	return met
}

type metrics struct {
	sessionId string

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
	receiverEstimatedTargetBitrate  prometheus.Gauge

	receiverReportDelay     prometheus.Gauge
	receiverReportJitter    prometheus.Gauge
	receiverReportTotalLost prometheus.Gauge

	iceBytesSent      prometheus.Gauge
	iceBytesReceived  prometheus.Gauge
	sctpBytesSent     prometheus.Gauge
	sctpBytesReceived prometheus.Gauge
}

func (met *metrics) reset() {
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

func (met *metrics) NewConnection() {
	met.connectionCount.Add(1)
}

func (met *metrics) NewICECandidate(candidate webrtc.ICECandidateStats) {
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

func (met *metrics) SetICECandidatesUsed(candidates []webrtc.ICECandidateStats) {
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

func (met *metrics) SetState(state webrtc.PeerConnectionState) {
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
		met.reset()
	default:
		met.connectionState.Set(-1)
	}

	met.connectionStateCount.Add(1)
}

func (met *metrics) SetVideoID(videoId string) {
	met.videoIdsMu.Lock()
	defer met.videoIdsMu.Unlock()

	if _, found := met.videoIds[videoId]; !found {
		met.videoIds[videoId] = promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "video_listeners",
			Namespace: "neko",
			Subsystem: "webrtc",
			Help:      "Listeners for Video pipelines by a session.",
			ConstLabels: map[string]string{
				"session_id": met.sessionId,
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

func (met *metrics) SetReceiverEstimatedMaximumBitrate(bitrate float32) {
	met.receiverEstimatedMaximumBitrate.Set(float64(bitrate))
}

func (met *metrics) SetReceiverEstimatedTargetBitrate(bitrate float64) {
	met.receiverEstimatedTargetBitrate.Set(bitrate)
}

func (met *metrics) SetReceiverReport(report rtcp.ReceptionReport) {
	met.receiverReportDelay.Set(float64(report.Delay))
	met.receiverReportJitter.Set(float64(report.Jitter))
	met.receiverReportTotalLost.Set(float64(report.TotalLost))
}

func (met *metrics) SetIceTransportStats(data webrtc.TransportStats) {
	met.iceBytesSent.Set(float64(data.BytesSent))
	met.iceBytesReceived.Set(float64(data.BytesReceived))
}

func (met *metrics) SetSctpTransportStats(data webrtc.TransportStats) {
	met.sctpBytesSent.Set(float64(data.BytesSent))
	met.sctpBytesReceived.Set(float64(data.BytesReceived))
}
