package capture

import (
	"sync"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/codec"
	"demodesk/neko/internal/capture/gst"
)

type StreamManagerCtx struct {
	logger          zerolog.Logger
	mu              sync.Mutex
	codec           codec.RTPCodec
	pipelineDevice  string
	pipelineSrc     string
	pipeline        *gst.Pipeline
	sample          chan types.Sample
	emmiter         events.EventEmmiter
	emitUpdate      chan bool
	emitStop        chan bool
	enabled         bool
}

func streamNew(codec codec.RTPCodec, pipelineDevice string, pipelineSrc string) *StreamManagerCtx {
	manager := &StreamManagerCtx{
		logger:          log.With().Str("module", "capture").Str("submodule", "stream").Logger(),
		mu:              sync.Mutex{},
		codec:           codec,
		pipelineDevice:  pipelineDevice,
		pipelineSrc:     pipelineSrc,
		emmiter:         events.New(),
		emitUpdate:      make(chan bool),
		emitStop:        make(chan bool),
		enabled:         false,
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

func (manager *StreamManagerCtx) Shutdown() {
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

func (manager *StreamManagerCtx) Start() {
	manager.enabled = true
	manager.createPipeline()
}

func (manager *StreamManagerCtx) Stop() {
	manager.enabled = false
	manager.destroyPipeline()
}

func (manager *StreamManagerCtx) Enabled() bool {
	return manager.enabled
}

func (manager *StreamManagerCtx) createPipeline() error {
	var err error

	codec := manager.Codec()
	manager.logger.Info().
		Str("codec", codec.Name).
		Str("device", manager.pipelineDevice).
		Str("src", manager.pipelineSrc).
		Msgf("creating pipeline")

	manager.pipeline, err = gst.CreateAppPipeline(codec, manager.pipelineDevice, manager.pipelineSrc)
	if err != nil {
		return err
	}

	manager.logger.Info().
		Str("src", manager.pipeline.Src).
		Msgf("starting pipeline")

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
	manager.logger.Info().Msgf("stopping pipeline")
	manager.pipeline = nil
}
