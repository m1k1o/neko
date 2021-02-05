package capture

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/capture/gst"
)

type BroacastManagerCtx struct {
	logger    zerolog.Logger
	mu        sync.Mutex
	config    *config.Capture
	pipeline  *gst.Pipeline
	enabled   bool
	url       string
}

func broadcastNew(config *config.Capture) *BroacastManagerCtx {
	return &BroacastManagerCtx{
		logger:   log.With().Str("module", "capture").Str("submodule", "broadcast").Logger(),
		mu:       sync.Mutex{},
		config:   config,
		enabled:  false,
		url:      "",
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
	manager.enabled = true
	return nil
}

func (manager *BroacastManagerCtx) Stop() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.enabled = false
	manager.destroyPipeline()
}

func (manager *BroacastManagerCtx) Enabled() bool {
	return manager.enabled
}

func (manager *BroacastManagerCtx) Url() string {
	return manager.url
}

func (manager *BroacastManagerCtx) createPipeline() error {
	if manager.pipeline != nil {
		return fmt.Errorf("pipeline already running")
	}

	var err error

	manager.logger.Info().
		Str("audio_device", manager.config.Device).
		Str("video_display", manager.config.Display).
		Str("broadcast_pipeline", manager.config.BroadcastPipeline).
		Msgf("creating pipeline")

	manager.pipeline, err = gst.CreateRTMPPipeline(
		manager.config.Device,
		manager.config.Display,
		manager.config.BroadcastPipeline,
		manager.url,
	)

	if err != nil {
		return err
	}

	manager.pipeline.Play()
	manager.logger.Info().Msgf("starting pipeline")
	return nil
}

func (manager *BroacastManagerCtx) destroyPipeline() {
	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Stop()
	manager.logger.Info().Msgf("stopping pipeline")
	manager.pipeline = nil
}
