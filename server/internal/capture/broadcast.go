package capture

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/pkg/gst"
	"github.com/demodesk/neko/pkg/types"
)

type BroacastManagerCtx struct {
	logger zerolog.Logger
	mu     sync.Mutex

	pipeline   gst.Pipeline
	pipelineMu sync.Mutex
	pipelineFn func(url string) (string, error)

	url     string
	started bool

	// metrics
	pipelinesCounter prometheus.Counter
	pipelinesActive  prometheus.Gauge
}

func broadcastNew(pipelineFn func(url string) (string, error), defaultUrl string) *BroacastManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "broadcast").
		Logger()

	return &BroacastManagerCtx{
		logger:     logger,
		pipelineFn: pipelineFn,
		url:        defaultUrl,
		started:    defaultUrl != "",

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

	manager.destroyPipeline()
}

func (manager *BroacastManagerCtx) Start(url string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	err := manager.createPipeline()
	if err != nil {
		return err
	}

	manager.url = url
	manager.started = true
	return nil
}

func (manager *BroacastManagerCtx) Stop() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.started = false
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

	manager.pipeline, err = gst.CreatePipeline(pipelineStr)
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
