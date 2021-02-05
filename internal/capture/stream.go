package capture

import (
	"fmt"
	"sync"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/codec"
	"demodesk/neko/internal/capture/gst"
)

type StreamManagerCtx struct {
	logger       zerolog.Logger
	mu           sync.Mutex
	codec        codec.RTPCodec
	pipelineStr  string
	pipeline     *gst.Pipeline
	sample       chan types.Sample
	emmiter      events.EventEmmiter
	emitUpdate   chan bool
	emitStop     chan bool
	enabled      bool
}

func streamNew(codec codec.RTPCodec, pipelineStr string) *StreamManagerCtx {
	manager := &StreamManagerCtx{
		logger:       log.With().Str("module", "capture").Str("submodule", "stream").Logger(),
		codec:        codec,
		pipelineStr:  pipelineStr,
		emmiter:      events.New(),
		emitUpdate:   make(chan bool),
		emitStop:     make(chan bool),
		enabled:      false,
	}

	go func() {
		manager.logger.Debug().Msg("started emitting samples")

		for {
			select {
			case <-manager.emitStop:
				manager.logger.Debug().Msg("stopped emitting samples")
				return
			case <-manager.emitUpdate:
				manager.logger.Debug().Msg("update emitting samples")
			case sample := <-manager.sample:
				manager.emmiter.Emit("sample", sample)
			}
		}
	}()

	return manager
}

func (manager *StreamManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutting down")

	manager.destroyPipeline()
	manager.emitStop <- true
}

func (manager *StreamManagerCtx) Codec() codec.RTPCodec {
	return manager.codec
}

func (manager *StreamManagerCtx) OnSample(listener func(sample types.Sample)) {
	manager.emmiter.On("sample", func(payload ...interface{}) {
		listener(payload[0].(types.Sample))
	})
}

func (manager *StreamManagerCtx) Start() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	err := manager.createPipeline()
	if err != nil {
		return err
	}

	manager.enabled = true
	return nil
}

func (manager *StreamManagerCtx) Stop() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.enabled = false
	manager.destroyPipeline()
}

func (manager *StreamManagerCtx) Enabled() bool {
	return manager.enabled
}

func (manager *StreamManagerCtx) createPipeline() error {
	if manager.pipeline != nil {
		return fmt.Errorf("pipeline already exists")
	}

	var err error

	codec := manager.Codec()
	manager.logger.Info().
		Str("codec", codec.Name).
		Str("src", manager.pipelineStr).
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

func (manager *StreamManagerCtx) destroyPipeline() {
	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Stop()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil
}
