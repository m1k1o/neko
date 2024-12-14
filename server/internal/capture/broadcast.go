package capture

import (
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/capture/gst"
	"m1k1o/neko/internal/types"
)

type PipelineSignal int

const (
	PL_STOP  PipelineSignal = 0x00
	PL_START                = 0x01
)

type BroacastManagerCtx struct {
	logger zerolog.Logger
	mu     sync.Mutex

	pipeline        *gst.Pipeline
	pipelineMu      sync.Mutex
	pipelineFn      func(url string) (string, error)
	pipelineRestart chan bool
	pipelineSignal  chan PipelineSignal

	url     string
	started bool
}

func broadcastNew(pipelineFn func(url string) (string, error), url string, started bool) *BroacastManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "broadcast").
		Logger()

	return &BroacastManagerCtx{
		logger:          logger,
		pipelineFn:      pipelineFn,
		pipelineRestart: make(chan bool),
		url:             url,
		started:         started && url != "",
	}
}

func (manager *BroacastManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.destroyPipeline()
}

func (manager *BroacastManagerCtx) Start(url string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.url = url

	err := manager.createPipeline()
	if err != nil {
		return err
	}

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

func (manager *BroacastManagerCtx) GetRestart() chan bool {
	return manager.pipelineRestart
}

func (manager *BroacastManagerCtx) GetSignal() chan PipelineSignal {
	return manager.pipelineSignal
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
	pipelineStr, err := manager.pipelineFn(manager.url)
	if err != nil {
		return err
	}

	manager.logger.Info().
		Str("url", manager.url).
		Str("src", pipelineStr).
		Msgf("starting pipeline")

	manager.pipeline, err = gst.CreatePipeline(pipelineStr)
	if err != nil {
		return err
	}

	manager.pipeline.Play()

	manager.logger.Info().Msgf("Pipeline started")

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
