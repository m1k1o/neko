package capture

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/gst"
	"github.com/m1k1o/neko/server/pkg/types"
)

type broadcastPipeline interface {
	Play()
	Destroy()
}

type BroacastManagerCtx struct {
	logger zerolog.Logger
	mu     sync.Mutex

	pipeline        broadcastPipeline
	pipelineMu      sync.Mutex
	pipelineFn      func(url string) (string, error)
	pipelineFactory func(string) (broadcastPipeline, error)

	url     string
	started bool

	// stop the broadcast once the room has been empty for this long, 0 disables it
	emptyTimeout time.Duration
	emptyTimer   *time.Timer
	roomEmpty    func() bool

	// metrics
	pipelinesCounter prometheus.Counter
	pipelinesActive  prometheus.Gauge
}

func broadcastNew(pipelineFn func(url string) (string, error), defaultUrl string, autostart bool, emptyTimeout time.Duration) *BroacastManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "broadcast").
		Logger()

	return &BroacastManagerCtx{
		logger:     logger,
		pipelineFn: pipelineFn,
		pipelineFactory: func(src string) (broadcastPipeline, error) {
			return gst.CreatePipeline(src)
		},
		url:          defaultUrl,
		started:      defaultUrl != "" && autostart,
		emptyTimeout: emptyTimeout,

		// metrics
		pipelinesCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "pipelines_total",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of created pipelines.",
			ConstLabels: map[string]string{
				"submodule":  "broadcast",
				"video_id":   "main",
				"codec_name": "-",
				"codec_type": "-",
			},
		}),
		pipelinesActive: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "pipelines_active",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of active pipelines.",
			ConstLabels: map[string]string{
				"submodule":  "broadcast",
				"video_id":   "main",
				"codec_name": "-",
				"codec_type": "-",
			},
		}),
	}
}

func (manager *BroacastManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.mu.Lock()
	manager.stopEmptyTimer()
	manager.mu.Unlock()

	manager.destroyPipeline()
}

func (manager *BroacastManagerCtx) Start(url string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.url = url

	err := manager.createPipeline()
	if err != nil {
		return err
	}

	manager.started = true
	manager.updateEmptyTimer()
	return nil
}

func (manager *BroacastManagerCtx) Stop() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.started = false
	manager.updateEmptyTimer()
	manager.destroyPipeline()
}

func (manager *BroacastManagerCtx) Started() bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.started
}

func (manager *BroacastManagerCtx) Url() string {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.url
}

func (manager *BroacastManagerCtx) createPipeline() error {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline != nil {
		return types.ErrCapturePipelineAlreadyExists
	}

	pipelineStr, err := manager.pipelineFn(manager.url)
	if err != nil {
		return err
	}

	manager.logger.Info().
		Str("url", manager.url).
		Str("src", pipelineStr).
		Msgf("starting pipeline")

	manager.pipeline, err = manager.pipelineFactory(pipelineStr)
	if err != nil {
		return err
	}

	manager.pipeline.Play()
	manager.pipelinesCounter.Inc()
	manager.pipelinesActive.Set(1)

	return nil
}

func (manager *BroacastManagerCtx) destroyPipeline() {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Destroy()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil

	manager.pipelinesActive.Set(0)
}

// watchRoom enables stopping the broadcast once it has been running for emptyTimeout without
// anyone in the room. roomEmpty reports whether the room is empty, and roomChanged must be
// called every time its answer may have changed.
func (manager *BroacastManagerCtx) watchRoom(roomEmpty func() bool) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.roomEmpty = roomEmpty
	manager.updateEmptyTimer()
}

func (manager *BroacastManagerCtx) roomChanged() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.updateEmptyTimer()
}

// updateEmptyTimer starts the countdown when the broadcast is running in an empty room and
// cancels it otherwise. A countdown that is already running is kept, so that it measures the
// whole time since the room was last left or the broadcast was started. Must be called with mu.
func (manager *BroacastManagerCtx) updateEmptyTimer() {
	if manager.emptyTimeout <= 0 || manager.roomEmpty == nil {
		return
	}

	if !manager.started || !manager.roomEmpty() {
		manager.stopEmptyTimer()
		return
	}

	if manager.emptyTimer != nil {
		return
	}

	var timer *time.Timer
	timer = time.AfterFunc(manager.emptyTimeout, func() {
		manager.mu.Lock()
		defer manager.mu.Unlock()

		// cancelled or replaced while waiting for the lock
		if manager.emptyTimer != timer {
			return
		}
		manager.emptyTimer = nil

		// nobody is connected, so there is no admin to notify about the stopped broadcast,
		// whoever connects next receives the current broadcast status
		manager.logger.Info().
			Dur("empty_timeout", manager.emptyTimeout).
			Msg("room is empty, stopping broadcast")

		manager.started = false
		manager.destroyPipeline()
	})
	manager.emptyTimer = timer
}

// stopEmptyTimer cancels the countdown, if there is one. Must be called with mu.
func (manager *BroacastManagerCtx) stopEmptyTimer() {
	if manager.emptyTimer != nil {
		manager.emptyTimer.Stop()
		manager.emptyTimer = nil
	}
}
