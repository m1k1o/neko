package capture

import (
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/capture/gst"
	"m1k1o/neko/internal/types"
)

type BroacastManagerCtx struct {
	logger zerolog.Logger
	mu     sync.Mutex

	pipeline   *gst.Pipeline
	pipelineMu sync.Mutex
	pipelineFn func(url string) (*gst.Pipeline, error)

	url     string
	started bool
}

func broadcastNew(pipelineFn func(url string) (*gst.Pipeline, error), url string, started bool) *BroacastManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "broadcast").
		Logger()

	return &BroacastManagerCtx{
		logger:     logger,
		pipelineFn: pipelineFn,
		url:        url,
		started:    started,
	}
}

func (manager *BroacastManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.destroyPipeline()
}

func (manager *BroacastManagerCtx) Start(url string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	err := manager.createPipeline()
	if err != nil {
		return err
	}

	manager.url = url
	manager.started = true
	return nil
}

func (manager *BroacastManagerCtx) Stop() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.started = false
	manager.destroyPipeline()
}

func (manager *BroacastManagerCtx) Started() bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.started
}

func (manager *BroacastManagerCtx) Url() string {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.url
}

func (manager *BroacastManagerCtx) createPipeline() error {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline != nil {
		return types.ErrCapturePipelineAlreadyExists
	}

	var err error

	manager.logger.Info().
		Str("url", manager.url).
		Msgf("creating pipeline")

	manager.pipeline, err = manager.pipelineFn(manager.url)
	if err != nil {
		return err
	}

	manager.logger.Info().
		Str("url", manager.url).
		Str("src", manager.pipeline.Src).
		Msgf("created pipeline")

	manager.pipeline.Play()

	return nil
}

func (manager *BroacastManagerCtx) destroyPipeline() {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Destroy()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil
}
