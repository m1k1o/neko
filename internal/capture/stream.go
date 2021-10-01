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

var moveListenerMu = sync.Mutex{}

type StreamManagerCtx struct {
	logger zerolog.Logger
	mu     sync.Mutex
	wg     sync.WaitGroup

	codec       codec.RTPCodec
	pipeline    *gst.Pipeline
	pipelineMu  sync.Mutex
	pipelineStr func() string

	sample       chan types.Sample
	sampleStop   chan interface{}
	sampleUpdate chan interface{}

	listeners   map[uintptr]*func(sample types.Sample)
	listenersMu sync.Mutex
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

func (manager *StreamManagerCtx) start() error {
	if len(manager.listeners) == 0 {
		err := manager.createPipeline()
		if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
			return err
		}

		manager.logger.Info().Msgf("first listener, starting")
	}

	return nil
}

func (manager *StreamManagerCtx) stop() {
	if len(manager.listeners) == 0 {
		manager.destroyPipeline()
		manager.logger.Info().Msgf("last listener, stopping")
	}
}

func (manager *StreamManagerCtx) addListener(listener *func(sample types.Sample)) {
	ptr := reflect.ValueOf(listener).Pointer()

	manager.listenersMu.Lock()
	manager.listeners[ptr] = listener
	manager.listenersMu.Unlock()

	manager.logger.Debug().Interface("ptr", ptr).Msgf("adding listener")
}

func (manager *StreamManagerCtx) removeListener(listener *func(sample types.Sample)) {
	ptr := reflect.ValueOf(listener).Pointer()

	manager.listenersMu.Lock()
	delete(manager.listeners, ptr)
	manager.listenersMu.Unlock()

	manager.logger.Debug().Interface("ptr", ptr).Msgf("removing listener")
}

func (manager *StreamManagerCtx) AddListener(listener *func(sample types.Sample)) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if listener == nil {
		return errors.New("listener cannot be nil")
	}

	// start if stopped
	if err := manager.start(); err != nil {
		return err
	}

	// add listener
	manager.addListener(listener)

	return nil
}

func (manager *StreamManagerCtx) RemoveListener(listener *func(sample types.Sample)) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if listener == nil {
		return errors.New("listener cannot be nil")
	}

	// remove listener
	manager.removeListener(listener)

	// stop if started
	manager.stop()

	return nil
}

// moving listeners between streams ensures, that target pipeline is running
// before listener is added, and stops source pipeline if there are 0 listeners
func (manager *StreamManagerCtx) MoveListenerTo(listener *func(sample types.Sample), stream types.StreamManager) error {
	if listener == nil {
		return errors.New("listener cannot be nil")
	}

	targetStream, ok := stream.(*StreamManagerCtx)
	if !ok {
		return errors.New("target stream manager does not support moving listeners")
	}

	// we need to acquire both mutextes, from source stream and from target stream
	// in order to do that safely (without possibility of deadlock) we need third
	// global mutex, that ensures atomic locking

	// lock global mutex
	moveListenerMu.Lock()

	// lock source stream
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// lock target stream
	targetStream.mu.Lock()
	defer targetStream.mu.Unlock()

	// unlock global mutex
	moveListenerMu.Unlock()

	// start if stopped
	if err := targetStream.start(); err != nil {
		return err
	}

	// swap listeners
	manager.removeListener(listener)
	targetStream.addListener(listener)

	// stop if started
	manager.stop()

	return nil
}

func (manager *StreamManagerCtx) ListenersCount() int {
	manager.listenersMu.Lock()
	defer manager.listenersMu.Unlock()

	return len(manager.listeners)
}

func (manager *StreamManagerCtx) Started() bool {
	return manager.ListenersCount() > 0
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
