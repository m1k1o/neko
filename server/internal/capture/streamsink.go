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

type StreamSinkManagerCtx struct {
	logger        zerolog.Logger
	mu            sync.Mutex
	sampleChannel chan types.Sample

	codec      codec.RTPCodec
	pipeline   *gst.Pipeline
	pipelineMu sync.Mutex
	pipelineFn func() (string, error)

	listeners   int
	listenersMu sync.Mutex
}

func streamSinkNew(codec codec.RTPCodec, pipelineFn func() (string, error), video_id string) *StreamSinkManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "stream-sink").
		Str("video_id", video_id).Logger()

	manager := &StreamSinkManagerCtx{
		logger:        logger,
		codec:         codec,
		pipelineFn:    pipelineFn,
		sampleChannel: make(chan types.Sample),
	}

	return manager
}

func (manager *StreamSinkManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.destroyPipeline()
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

	pipelineStr, err := manager.pipelineFn()
	if err != nil {
		return err
	}

	manager.logger.Info().
		Str("codec", manager.codec.Name).
		Str("src", pipelineStr).
		Msgf("creating pipeline")

	manager.pipeline, err = gst.CreatePipeline(pipelineStr)
	if err != nil {
		return err
	}

	appsinkSubfix := "audio"
	if manager.codec.IsVideo() {
		appsinkSubfix = "video"
	}

	manager.pipeline.AttachAppsink("appsink"+appsinkSubfix, manager.sampleChannel)
	manager.pipeline.Play()

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

func (manager *StreamSinkManagerCtx) GetSampleChannel() chan types.Sample {
	return manager.sampleChannel
}
