package capture

import (
	"fmt"
	"time"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/capture/gst"
)

type ScreencastManagerCtx struct {
	logger    zerolog.Logger
	mu        sync.Mutex
	config    *config.Capture
	pipeline  *gst.Pipeline
	enabled   bool
	started   bool
	shutdown  chan bool
	refresh   chan bool
	expires   time.Time
	timeout   time.Duration
	sample    chan types.Sample
	image     types.Sample
}

func screencastNew(config *config.Capture) *ScreencastManagerCtx {
	manager := &ScreencastManagerCtx{
		logger:    log.With().Str("module", "capture").Str("submodule", "screencast").Logger(),
		config:    config,
		enabled:   config.Screencast,
		started:   false,
		shutdown:  make(chan bool),
		refresh:   make(chan bool),
		timeout:   10 * time.Second,
	}

	if !manager.enabled {
		return manager
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		manager.logger.Debug().Msg("subroutine started")

		for {
			select {
			case <-manager.shutdown:
				manager.logger.Debug().Msg("shutting down")
				ticker.Stop()
				return
			case <-manager.refresh:
				manager.logger.Debug().Msg("subroutine updated")
			case sample := <-manager.sample:
				manager.image = sample
			case <-ticker.C:
				if manager.started && time.Now().After(manager.expires) {
					manager.stop()
				}
			}
		}
	}()

	return manager
}

func (manager *ScreencastManagerCtx) Enabled() bool {
	return manager.enabled
}

func (manager *ScreencastManagerCtx) Started() bool {
	return manager.started
}

func (manager *ScreencastManagerCtx) Image() ([]byte, error) {
	manager.expires = time.Now().Add(manager.timeout)

	if !manager.started {
		err := manager.start()
		if err != nil {
			return nil, err
		}

		select {
		case sample := <-manager.sample:
			return sample.Data, nil
		case <-time.After(1 * time.Second):
			return nil, fmt.Errorf("timeouted")
		}
	}

	if manager.image.Data == nil {
		return nil, fmt.Errorf("image sample not found")
	}

	return manager.image.Data, nil
}

func (manager *ScreencastManagerCtx) start() error {
	if !manager.enabled {
		return fmt.Errorf("screenshot pipeline not enabled")
	}

	err := manager.createPipeline()
	if err != nil {
		return err
	}

	manager.started = true
	return nil
}

func (manager *ScreencastManagerCtx) stop() {
	manager.started = false
	manager.destroyPipeline()
}

func (manager *ScreencastManagerCtx) createPipeline() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	var err error

	manager.logger.Info().
		Str("video_display", manager.config.Display).
		Str("screencast_pipeline", manager.config.ScreencastPipeline).
		Msgf("creating pipeline")
	
	manager.pipeline, err = gst.CreateJPEGPipeline(
		manager.config.Display,
		manager.config.ScreencastPipeline,
		manager.config.ScreencastRate,
		manager.config.ScreencastQuality,
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
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Stop()
	manager.logger.Info().Msgf("stopping pipeline")
	manager.pipeline = nil
}
