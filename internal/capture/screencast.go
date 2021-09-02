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

type ScreencastManagerCtx struct {
	logger      zerolog.Logger
	mu          sync.Mutex
	wg          sync.WaitGroup
	pipelineStr string
	pipeline    *gst.Pipeline
	enabled     bool
	started     bool
	emitStop    chan bool
	emitUpdate  chan bool
	expired     int32
	sample      chan types.Sample
	image       types.Sample
}

// timeout between intervals, when screencast pipeline is checked
const screencastTimeout = 5 * time.Second

func screencastNew(enabled bool, pipelineStr string) *ScreencastManagerCtx {
	manager := &ScreencastManagerCtx{
		logger:      log.With().Str("module", "capture").Str("submodule", "screencast").Logger(),
		pipelineStr: pipelineStr,
		enabled:     enabled,
		started:     false,
		emitStop:    make(chan bool),
		emitUpdate:  make(chan bool),
	}

	manager.wg.Add(1)

	go func() {
		manager.logger.Debug().Msg("started emitting samples")
		defer manager.wg.Done()

		ticker := time.NewTicker(screencastTimeout)
		defer ticker.Stop()

		for {
			select {
			case <-manager.emitStop:
				manager.logger.Debug().Msg("stopped emitting samples")
				return
			case <-manager.emitUpdate:
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

	manager.emitStop <- true
	manager.wg.Wait()
}

func (manager *ScreencastManagerCtx) Enabled() bool {
	return manager.enabled
}

func (manager *ScreencastManagerCtx) Started() bool {
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

	manager.pipeline.Start()
	manager.sample = manager.pipeline.Sample
	manager.emitUpdate <- true
	return nil
}

func (manager *ScreencastManagerCtx) destroyPipeline() {
	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Stop()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil
}
