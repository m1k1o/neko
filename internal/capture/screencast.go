package capture

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/pkg/gst"
	"github.com/demodesk/neko/pkg/types"
)

// timeout between intervals, when screencast pipeline is checked
const screencastTimeout = 5 * time.Second

type ScreencastManagerCtx struct {
	logger zerolog.Logger
	mu     sync.Mutex
	wg     sync.WaitGroup

	pipeline    *gst.Pipeline
	pipelineStr string
	pipelineMu  sync.Mutex

	image      types.Sample
	imageMu    sync.Mutex
	tickerStop chan struct{}

	enabled bool
	started bool
	expired int32

	// metrics
	imagesCounter    prometheus.Counter
	pipelinesCounter prometheus.Counter
	pipelinesActive  prometheus.Gauge
}

func screencastNew(enabled bool, pipelineStr string) *ScreencastManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "screencast").
		Logger()

	manager := &ScreencastManagerCtx{
		logger:      logger,
		pipelineStr: pipelineStr,
		tickerStop:  make(chan struct{}),
		enabled:     enabled,
		started:     false,

		// metrics
		imagesCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "screencast_images_total",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of created images.",
		}),
		pipelinesCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "pipelines_total",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of created pipelines.",
			ConstLabels: map[string]string{
				"submodule":  "screencast",
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
				"submodule":  "screencast",
				"video_id":   "main",
				"codec_name": "-",
				"codec_type": "-",
			},
		}),
	}

	manager.wg.Add(1)

	go func() {
		defer manager.wg.Done()

		ticker := time.NewTicker(screencastTimeout)
		defer ticker.Stop()

		for {
			select {
			case <-manager.tickerStop:
				return
			case <-ticker.C:
				if manager.Started() && !atomic.CompareAndSwapInt32(&manager.expired, 0, 1) {
					manager.stop()
				}
			}
		}
	}()

	return manager
}

func (manager *ScreencastManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.destroyPipeline()

	close(manager.tickerStop)
	manager.wg.Wait()
}

func (manager *ScreencastManagerCtx) Enabled() bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.enabled
}

func (manager *ScreencastManagerCtx) Started() bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.started
}

func (manager *ScreencastManagerCtx) Image() ([]byte, error) {
	atomic.StoreInt32(&manager.expired, 0)

	err := manager.start()
	if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
		return nil, err
	}

	manager.imageMu.Lock()
	defer manager.imageMu.Unlock()

	if manager.image.Data == nil {
		return nil, errors.New("image data not found")
	}

	return manager.image.Data, nil
}

func (manager *ScreencastManagerCtx) start() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if !manager.enabled {
		return errors.New("screencast not enabled")
	}

	err := manager.createPipeline()
	if err != nil {
		return err
	}

	manager.started = true
	return nil
}

func (manager *ScreencastManagerCtx) stop() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.started = false
	manager.destroyPipeline()
}

func (manager *ScreencastManagerCtx) createPipeline() error {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline != nil {
		return types.ErrCapturePipelineAlreadyExists
	}

	var err error

	manager.logger.Info().
		Str("str", manager.pipelineStr).
		Msgf("creating pipeline")

	manager.pipeline, err = gst.CreatePipeline(manager.pipelineStr)
	if err != nil {
		return err
	}

	manager.pipeline.AttachAppsink("appsink")
	manager.pipeline.Play()
	manager.pipelinesCounter.Inc()
	manager.pipelinesActive.Set(1)

	// get first image
	select {
	case image, ok := <-manager.pipeline.Sample:
		if !ok {
			return errors.New("unable to get first image")
		} else {
			manager.setImage(image)
		}
	case <-time.After(1 * time.Second):
		return errors.New("timeouted while waiting for first image")
	}

	manager.wg.Add(1)
	pipeline := manager.pipeline

	go func() {
		manager.logger.Debug().Msg("started receiving images")
		defer manager.wg.Done()

		for {
			image, ok := <-pipeline.Sample
			if !ok {
				manager.logger.Debug().Msg("stopped receiving images")
				return
			}

			manager.setImage(image)
		}
	}()

	return nil
}

func (manager *ScreencastManagerCtx) setImage(image types.Sample) {
	manager.imageMu.Lock()
	manager.image = image
	manager.imageMu.Unlock()

	manager.imagesCounter.Inc()
}

func (manager *ScreencastManagerCtx) destroyPipeline() {
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
