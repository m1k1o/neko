package capture

import (
	"fmt"
	"sync"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/capture/gst"
)

type BroacastManagerCtx struct {
	logger       zerolog.Logger
	mu           sync.Mutex
	pipelineStr  string
	pipeline     *gst.Pipeline
	started      bool
	url          string
}

func broadcastNew(pipelineStr string) *BroacastManagerCtx {
	return &BroacastManagerCtx{
		logger:       log.With().Str("module", "capture").Str("submodule", "broadcast").Logger(),
		pipelineStr:  pipelineStr,
		started:      false,
		url:          "",
	}
}

func (manager *BroacastManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutting down")

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
	return manager.started
}

func (manager *BroacastManagerCtx) Url() string {
	return manager.url
}

func (manager *BroacastManagerCtx) createPipeline() error {
	if manager.pipeline != nil {
		return fmt.Errorf("pipeline already exists")
	}

	var err error

	// replace {url} with valid URL
	pipelineStr := strings.Replace(manager.pipelineStr, "{url}", manager.url, 1)

	manager.logger.Info().
		Str("str", pipelineStr).
		Msgf("starting pipeline")

	manager.pipeline, err = gst.CreatePipeline(pipelineStr)
	if err != nil {
		return err
	}

	manager.pipeline.Play()
	return nil
}

func (manager *BroacastManagerCtx) destroyPipeline() {
	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Stop()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil
}
