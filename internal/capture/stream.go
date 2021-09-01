package capture

import (
	"reflect"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/capture/gst"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/codec"
)

type StreamManagerCtx struct {
	logger      zerolog.Logger
	mu          sync.Mutex
	wg          sync.WaitGroup
	codec       codec.RTPCodec
	pipelineStr func() string
	pipeline    *gst.Pipeline
	sample      chan types.Sample
	listeners   map[uintptr]*func(sample types.Sample)
	emitMu      sync.Mutex
	emitUpdate  chan bool
	emitStop    chan bool
	started     bool
}

func streamNew(codec codec.RTPCodec, pipelineStr func() string, video_id string) *StreamManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "stream").
		Str("video_id", video_id).Logger()

	manager := &StreamManagerCtx{
		logger:      logger,
		codec:       codec,
		pipelineStr: pipelineStr,
		listeners:   map[uintptr]*func(sample types.Sample){},
		emitUpdate:  make(chan bool),
		emitStop:    make(chan bool),
		started:     false,
	}

	manager.wg.Add(1)

	go func() {
		manager.logger.Debug().Msg("started emitting samples")
		defer manager.wg.Done()

		for {
			select {
			case <-manager.emitStop:
				manager.logger.Debug().Msg("stopped emitting samples")
				return
			case <-manager.emitUpdate:
				manager.logger.Debug().Msg("update emitting samples")
			case sample := <-manager.sample:
				manager.emitMu.Lock()
				for _, emit := range manager.listeners {
					(*emit)(sample)
				}
				manager.emitMu.Unlock()
			}
		}
	}()

	return manager
}

func (manager *StreamManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.emitMu.Lock()
	for key := range manager.listeners {
		delete(manager.listeners, key)
	}
	manager.emitMu.Unlock()

	manager.destroyPipeline()

	manager.emitStop <- true
	manager.wg.Wait()
}

func (manager *StreamManagerCtx) Codec() codec.RTPCodec {
	return manager.codec
}

func (manager *StreamManagerCtx) AddListener(listener *func(sample types.Sample)) {
	manager.emitMu.Lock()
	defer manager.emitMu.Unlock()

	if listener != nil {
		ptr := reflect.ValueOf(listener).Pointer()
		manager.listeners[ptr] = listener
		manager.logger.Debug().Interface("ptr", ptr).Msgf("adding listener")
	}
}

func (manager *StreamManagerCtx) RemoveListener(listener *func(sample types.Sample)) {
	manager.emitMu.Lock()
	defer manager.emitMu.Unlock()

	if listener != nil {
		ptr := reflect.ValueOf(listener).Pointer()
		delete(manager.listeners, ptr)
		manager.logger.Debug().Interface("ptr", ptr).Msgf("removing listener")
	}
}

func (manager *StreamManagerCtx) ListenersCount() int {
	manager.emitMu.Lock()
	defer manager.emitMu.Unlock()

	return len(manager.listeners)
}

func (manager *StreamManagerCtx) Start() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	err := manager.createPipeline()
	if err != nil {
		return err
	}

	manager.logger.Info().Msgf("start")
	manager.started = true
	return nil
}

func (manager *StreamManagerCtx) Stop() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.logger.Info().Msgf("stop")
	manager.started = false
	manager.destroyPipeline()
}

func (manager *StreamManagerCtx) Started() bool {
	return manager.started
}

func (manager *StreamManagerCtx) createPipeline() error {
	if manager.pipeline != nil {
		return types.ErrCapturePipelineAlreadyExists
	}

	var err error

	codec := manager.Codec()
	pipelineStr := manager.pipelineStr()
	manager.logger.Info().
		Str("codec", codec.Name).
		Str("src", pipelineStr).
		Msgf("creating pipeline")

	manager.pipeline, err = gst.CreatePipeline(pipelineStr)
	if err != nil {
		return err
	}

	manager.pipeline.Start()

	manager.sample = manager.pipeline.Sample
	manager.emitUpdate <- true
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
