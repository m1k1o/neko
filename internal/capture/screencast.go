package capture

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/capture/gst"
	"demodesk/neko/internal/types"
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

	image        types.Sample
	sample       chan types.Sample
	sampleStop   chan struct{}
	sampleUpdate chan struct{}

	enabled bool
	started bool
	expired int32
}

func screencastNew(enabled bool, pipelineStr string) *ScreencastManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "screencast").
		Logger()

	manager := &ScreencastManagerCtx{
		logger:       logger,
		pipelineStr:  pipelineStr,
		sampleStop:   make(chan struct{}),
		sampleUpdate: make(chan struct{}),
		enabled:      enabled,
		started:      false,
	}

	manager.wg.Add(1)

	go func() {
		manager.logger.Debug().Msg("started emitting samples")
		defer manager.wg.Done()

		ticker := time.NewTicker(screencastTimeout)
		defer ticker.Stop()

		for {
			select {
			case <-manager.sampleStop:
				manager.logger.Debug().Msg("stopped emitting samples")
				return
			case <-manager.sampleUpdate:
				manager.logger.Debug().Msg("update emitting samples")
			case sample := <-manager.sample:
				manager.image = sample
			case <-ticker.C:
				if manager.started && !atomic.CompareAndSwapInt32(&manager.expired, 0, 1) {
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

	close(manager.sampleStop)
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

	if !manager.started {
		err := manager.start()
		if err != nil {
			return nil, err
		}

		select {
		case sample := <-manager.sample:
			return sample.Data, nil
		case <-time.After(1 * time.Second):
			return nil, errors.New("timeouted")
		}
	}

	if manager.image.Data == nil {
		return nil, errors.New("image sample not found")
	}

	return manager.image.Data, nil
}

func (manager *ScreencastManagerCtx) start() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if !manager.enabled {
		return errors.New("screenshot pipeline not enabled")
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

	manager.sample = manager.pipeline.Sample
	manager.sampleUpdate <- struct{}{}
	return nil
}

func (manager *ScreencastManagerCtx) destroyPipeline() {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Stop()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil
}
