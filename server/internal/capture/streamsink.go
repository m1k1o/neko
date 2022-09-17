package capture

import (
	"errors"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/capture/gst"
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/codec"
)

var moveSinkListenerMu = sync.Mutex{}

type StreamSinkManagerCtx struct {
	logger zerolog.Logger
	mu     sync.Mutex
	wg     sync.WaitGroup

	codec      codec.RTPCodec
	pipeline   *gst.Pipeline
	pipelineMu sync.Mutex
	pipelineFn func() (*gst.Pipeline, error)

	listeners   int
	listenersMu sync.Mutex

	sampleFn func(sample types.Sample)
}

func streamSinkNew(codec codec.RTPCodec, pipelineFn func() (*gst.Pipeline, error), video_id string) *StreamSinkManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "stream-sink").
		Str("video_id", video_id).Logger()

	manager := &StreamSinkManagerCtx{
		logger:     logger,
		codec:      codec,
		pipelineFn: pipelineFn,
	}

	return manager
}

func (manager *StreamSinkManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.destroyPipeline()
	manager.wg.Wait()
}

func (manager *StreamSinkManagerCtx) OnSample(listener func(sample types.Sample)) {
	manager.sampleFn = listener
}

func (manager *StreamSinkManagerCtx) Codec() codec.RTPCodec {
	return manager.codec
}

func (manager *StreamSinkManagerCtx) start() error {
	if manager.listeners == 0 {
		err := manager.createPipeline()
		if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
			return err
		}

		manager.logger.Info().Msgf("first listener, starting")
	}

	return nil
}

func (manager *StreamSinkManagerCtx) stop() {
	if manager.listeners == 0 {
		manager.destroyPipeline()
		manager.logger.Info().Msgf("last listener, stopping")
	}
}

func (manager *StreamSinkManagerCtx) addListener() {
	manager.listenersMu.Lock()
	manager.listeners++
	manager.listenersMu.Unlock()
}

func (manager *StreamSinkManagerCtx) removeListener() {
	manager.listenersMu.Lock()
	manager.listeners--
	manager.listenersMu.Unlock()
}

func (manager *StreamSinkManagerCtx) AddListener() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// start if stopped
	if err := manager.start(); err != nil {
		return err
	}

	// add listener
	manager.addListener()

	return nil
}

func (manager *StreamSinkManagerCtx) RemoveListener() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// remove listener
	manager.removeListener()

	// stop if started
	manager.stop()

	return nil
}

func (manager *StreamSinkManagerCtx) ListenersCount() int {
	manager.listenersMu.Lock()
	defer manager.listenersMu.Unlock()

	return manager.listeners
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

	manager.logger.Info().
		Str("codec", manager.codec.Name).
		Msgf("creating pipeline")

	manager.pipeline, err = manager.pipelineFn()
	if err != nil {
		return err
	}

	manager.logger.Info().
		Str("codec", manager.codec.Name).
		Str("src", manager.pipeline.Src).
		Msgf("created pipeline")

	manager.pipeline.AttachAppsink("appsink")
	manager.pipeline.Play()

	manager.wg.Add(1)
	pipeline := manager.pipeline

	go func() {
		manager.logger.Debug().Msg("started emitting samples")
		defer manager.wg.Done()

		for {
			sample, ok := <-pipeline.Sample
			if !ok {
				manager.logger.Debug().Msg("stopped emitting samples")
				return
			}

			manager.sampleFn(sample)
		}
	}()

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
