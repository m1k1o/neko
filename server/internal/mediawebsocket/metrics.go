package mediawebsocket

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics for tuning the backpressure window. These are package level and
// labelled only by track, so cardinality does not grow with sessions.
var (
	droppedSamples = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "dropped_samples_total",
		Namespace: "neko",
		Subsystem: "mediawebsocket",
		Help:      "Samples dropped before delivery because a client fell behind.",
	}, []string{"track", "reason"})

	keyframeRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "keyframe_requests_total",
		Namespace: "neko",
		Subsystem: "mediawebsocket",
		Help:      "Keyframe requests made to recover a broken decode chain.",
	})

	// unackedFrames records the in-flight window depth at send time. This is
	// the primary signal for tuning maxUnackedFrames.
	unackedFrames = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:      "unacked_frames",
		Namespace: "neko",
		Subsystem: "mediawebsocket",
		Help:      "Number of unacked video frames in flight when a frame is sent.",
		Buckets:   prometheus.LinearBuckets(0, 1, maxUnackedFrames+2),
	})

	audioQueueDepth = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:      "audio_queue_depth",
		Namespace: "neko",
		Subsystem: "mediawebsocket",
		Help:      "Buffered audio samples awaiting delivery when a sample is queued.",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 7),
	})
)
