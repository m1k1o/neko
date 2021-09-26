package capture

import (
	"errors"
	"reflect"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/capture/gst"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/codec"
)

type StreamManagerCtx struct {
	logger zerolog.Logger
	mu     sync.Mutex
	wg     sync.WaitGroup
	codec  codec.RTPCodec

	pipeline    *gst.Pipeline
	pipelineMu  sync.Mutex
	pipelineStr func() string

	sample       chan types.Sample
	sampleStop   chan interface{}
	sampleUpdate chan interface{}

	listeners      map[uintptr]*func(sample types.Sample)
	listenersMu    sync.Mutex
	listenersCount uint32
}

func streamNew(codec codec.RTPCodec, pipelineStr func() string, video_id string) *StreamManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "stream").
		Str("video_id", video_id).Logger()

	manager := &StreamManagerCtx{
		logger:       logger,
		codec:        codec,
		pipelineStr:  pipelineStr,
		sampleStop:   make(chan interface{}),
		sampleUpdate: make(chan interface{}),
		listeners:    map[uintptr]*func(sample types.Sample){},
	}

	manager.wg.Add(1)

	go func() {
		manager.logger.Debug().Msg("started emitting samples")
		defer manager.wg.Done()

		for {
			select {
			case <-manager.sampleStop:
				manager.logger.Debug().Msg("stopped emitting samples")
				return
			case <-manager.sampleUpdate:
				manager.logger.Debug().Msg("update emitting samples")
			case sample := <-manager.sample:
				manager.listenersMu.Lock()
				for _, emit := range manager.listeners {
					(*emit)(sample)
				}
				manager.listenersMu.Unlock()
			}
		}
	}()

	return manager
}

func (manager *StreamManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.listenersMu.Lock()
	for key := range manager.listeners {
		delete(manager.listeners, key)
	}
	manager.listenersMu.Unlock()

	manager.destroyPipeline()

	close(manager.sampleStop)
	manager.wg.Wait()
}

func (manager *StreamManagerCtx) Codec() codec.RTPCodec {
	return manager.codec
}

func (manager *StreamManagerCtx) NewListener(listener *func(sample types.Sample)) (addListener func(), err error) {
	if listener == nil {
		return addListener, errors.New("listener cannot be nil")
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.listenersCount++
	if manager.listenersCount == 1 {
		err := manager.createPipeline()
		if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
			return addListener, err
		}

		manager.logger.Info().Msgf("first listener, starting")
	}

	return func() {
		ptr := reflect.ValueOf(listener).Pointer()

		manager.listenersMu.Lock()
		manager.listeners[ptr] = listener
		manager.listenersMu.Unlock()

		manager.logger.Debug().Interface("ptr", ptr).Msgf("adding listener")
	}, nil
}

func (manager *StreamManagerCtx) RemoveListener(listener *func(sample types.Sample)) {
	if listener == nil {
		return
	}

	ptr := reflect.ValueOf(listener).Pointer()

	manager.listenersMu.Lock()
	delete(manager.listeners, ptr)
	manager.listenersMu.Unlock()

	manager.logger.Debug().Interface("ptr", ptr).Msgf("removing listener")

	manager.mu.Lock()
	manager.listenersCount--
	manager.mu.Unlock()

	go func() {
		manager.mu.Lock()
		defer manager.mu.Unlock()

		if manager.listenersCount == 0 {
			manager.destroyPipeline()
			manager.listenersCount = 0
			manager.logger.Info().Msgf("last listener, stopping")
		}
	}()
}

func (manager *StreamManagerCtx) ListenersCount() int {
	manager.listenersMu.Lock()
	defer manager.listenersMu.Unlock()

	return len(manager.listeners)
}

func (manager *StreamManagerCtx) Started() bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.listenersCount > 0
}

func (manager *StreamManagerCtx) createPipeline() error {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

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
	manager.sampleUpdate <- struct{}{}
	return nil
}

func (manager *StreamManagerCtx) destroyPipeline() {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Stop()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil
}
