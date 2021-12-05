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

var moveSinkListenerMu = sync.Mutex{}

type StreamSinkManagerCtx struct {
	logger zerolog.Logger
	mu     sync.Mutex
	wg     sync.WaitGroup

	codec       codec.RTPCodec
	pipeline    *gst.Pipeline
	pipelineMu  sync.Mutex
	pipelineStr func() string

	sample       chan types.Sample
	sampleStop   chan struct{}
	sampleUpdate chan struct{}

	listeners   map[uintptr]*func(sample types.Sample)
	listenersMu sync.Mutex
}

func streamSinkNew(codec codec.RTPCodec, pipelineStr func() string, video_id string) *StreamSinkManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "stream-sink").
		Str("video_id", video_id).Logger()

	manager := &StreamSinkManagerCtx{
		logger:       logger,
		codec:        codec,
		pipelineStr:  pipelineStr,
		sampleStop:   make(chan struct{}),
		sampleUpdate: make(chan struct{}),
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

func (manager *StreamSinkManagerCtx) shutdown() {
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

func (manager *StreamSinkManagerCtx) Codec() codec.RTPCodec {
	return manager.codec
}

func (manager *StreamSinkManagerCtx) start() error {
	if len(manager.listeners) == 0 {
		err := manager.createPipeline()
		if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
			return err
		}

		manager.logger.Info().Msgf("first listener, starting")
	}

	return nil
}

func (manager *StreamSinkManagerCtx) stop() {
	if len(manager.listeners) == 0 {
		manager.destroyPipeline()
		manager.logger.Info().Msgf("last listener, stopping")
	}
}

func (manager *StreamSinkManagerCtx) addListener(listener *func(sample types.Sample)) {
	ptr := reflect.ValueOf(listener).Pointer()

	manager.listenersMu.Lock()
	manager.listeners[ptr] = listener
	manager.listenersMu.Unlock()

	manager.logger.Debug().Interface("ptr", ptr).Msgf("adding listener")
}

func (manager *StreamSinkManagerCtx) removeListener(listener *func(sample types.Sample)) {
	ptr := reflect.ValueOf(listener).Pointer()

	manager.listenersMu.Lock()
	delete(manager.listeners, ptr)
	manager.listenersMu.Unlock()

	manager.logger.Debug().Interface("ptr", ptr).Msgf("removing listener")
}

func (manager *StreamSinkManagerCtx) AddListener(listener *func(sample types.Sample)) error {
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

func (manager *StreamSinkManagerCtx) RemoveListener(listener *func(sample types.Sample)) error {
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
func (manager *StreamSinkManagerCtx) MoveListenerTo(listener *func(sample types.Sample), stream types.StreamSinkManager) error {
	if listener == nil {
		return errors.New("listener cannot be nil")
	}

	targetStream, ok := stream.(*StreamSinkManagerCtx)
	if !ok {
		return errors.New("target stream manager does not support moving listeners")
	}

	// we need to acquire both mutextes, from source stream and from target stream
	// in order to do that safely (without possibility of deadlock) we need third
	// global mutex, that ensures atomic locking

	// lock global mutex
	moveSinkListenerMu.Lock()

	// lock source stream
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// lock target stream
	targetStream.mu.Lock()
	defer targetStream.mu.Unlock()

	// unlock global mutex
	moveSinkListenerMu.Unlock()

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

func (manager *StreamSinkManagerCtx) ListenersCount() int {
	manager.listenersMu.Lock()
	defer manager.listenersMu.Unlock()

	return len(manager.listeners)
}

func (manager *StreamSinkManagerCtx) Started() bool {
	return manager.ListenersCount() > 0
}

func (manager *StreamSinkManagerCtx) createPipeline() error {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline != nil {
		return types.ErrCapturePipelineAlreadyExists
	}

	var err error

	pipelineStr := manager.pipelineStr()
	manager.logger.Info().
		Str("codec", manager.codec.Name).
		Str("src", pipelineStr).
		Msgf("creating pipeline")

	manager.pipeline, err = gst.CreatePipeline(pipelineStr)
	if err != nil {
		return err
	}

	manager.pipeline.AttachAppsink("appsink")
	manager.pipeline.Play()

	manager.sample = manager.pipeline.Sample
	manager.sampleUpdate <- struct{}{}
	return nil
}

func (manager *StreamSinkManagerCtx) destroyPipeline() {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Destroy()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil
}
