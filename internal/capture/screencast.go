package capture

import (
	"bytes"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/capture/gst"
)

type ScreencastManagerCtx struct {
	logger    zerolog.Logger
	config    *config.Capture
	pipeline  *gst.Pipeline
	enabled   bool
	shutdown  chan bool
	refresh   chan bool
	sample    chan types.Sample
	image     *bytes.Buffer
}

func screencastNew(config *config.Capture) *ScreencastManagerCtx {
	manager := &ScreencastManagerCtx{
		logger:    log.With().Str("module", "capture").Str("submodule", "screencast").Logger(),
		config:    config,
		enabled:   config.Screencast,
		shutdown:  make(chan bool),
		refresh:   make(chan bool),
		image:     new(bytes.Buffer),
	}

	go func() {
		manager.logger.Debug().Msg("subroutine started")

		for {
			select {
			case <-manager.shutdown:
				manager.logger.Debug().Msg("shutting down")
				return
			case <-manager.refresh:
				manager.logger.Debug().Msg("subroutine updated")
			case sample := <-manager.sample:
				manager.image.Reset()
				manager.image.Write(sample.Data)
			}
		}
	}()

	return manager
}

func (manager *ScreencastManagerCtx) Start() error {
	manager.enabled = true
	return manager.createPipeline()
}

func (manager *ScreencastManagerCtx) Stop() {
	manager.enabled = false
	manager.destroyPipeline()
}

func (manager *ScreencastManagerCtx) Enabled() bool {
	return manager.enabled
}

func (manager *ScreencastManagerCtx) Image() []byte {
	return manager.image.Bytes()
}

func (manager *ScreencastManagerCtx) createPipeline() error {
	var err error

	manager.logger.Info().
		Str("video_display", manager.config.Display).
		Str("screencast_pipeline", manager.config.ScreencastPipeline).
		Msgf("creating pipeline")
	
	manager.pipeline, err = gst.CreateJPEGPipeline(
		manager.config.Display,
		manager.config.ScreencastPipeline,
	)

	if err != nil {
		return err
	}

	manager.pipeline.Start()
	manager.logger.Info().Msgf("starting pipeline")

	manager.sample = manager.pipeline.Sample
	manager.refresh <-true
	return nil
}

func (manager *ScreencastManagerCtx) destroyPipeline() {
	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Stop()
	manager.logger.Info().Msgf("stopping pipeline")
	manager.pipeline = nil
}
