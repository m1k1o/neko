package capture

import (
	"fmt"
	"time"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/capture/gst"
)

type ScreencastManagerCtx struct {
	logger       zerolog.Logger
	mu           sync.Mutex
	pipelineStr  string
	pipeline     *gst.Pipeline
	enabled      bool
	started      bool
	emitStop     chan bool
	emitUpdate   chan bool
	expired      int32
	sample       chan types.Sample
	image        types.Sample
}

const screencastTimeout = 5 * time.Second

func screencastNew(enabled bool, pipelineStr string) *ScreencastManagerCtx {
	manager := &ScreencastManagerCtx{
		logger:       log.With().Str("module", "capture").Str("submodule", "screencast").Logger(),
		pipelineStr:  pipelineStr,
		enabled:      enabled,
		started:      false,
		emitStop:     make(chan bool),
		emitUpdate:   make(chan bool),
	}

	go func() {
		ticker := time.NewTicker(screencastTimeout)
		manager.logger.Debug().Msg("started emitting samples")

		for {
			select {
			case <-manager.emitStop:
				manager.logger.Debug().Msg("stopped emitting samples")
				ticker.Stop()
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
	manager.logger.Info().Msgf("shutting down")

	manager.destroyPipeline()
	manager.emitStop <- true
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
			return nil, fmt.Errorf("timeouted")
		}
	}

	if manager.image.Data == nil {
		return nil, fmt.Errorf("image sample not found")
	}

	return manager.image.Data, nil
}

func (manager *ScreencastManagerCtx) start() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if !manager.enabled {
		return fmt.Errorf("screenshot pipeline not enabled")
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
		return fmt.Errorf("pipeline already exists")
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
	manager.emitUpdate <-true
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
