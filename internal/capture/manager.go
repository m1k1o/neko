package capture

import (
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/config"
)

type CaptureManagerCtx struct {
	logger       zerolog.Logger
	mu           sync.Mutex
	config       *config.Capture
	desktop      types.DesktopManager
	streaming    bool
	broadcast    *BroacastManagerCtx
	screencast   *ScreencastManagerCtx
	audio        *StreamManagerCtx
	video        *StreamManagerCtx
}

func New(desktop types.DesktopManager, config *config.Capture) *CaptureManagerCtx {
	return &CaptureManagerCtx{
		logger:      log.With().Str("module", "capture").Logger(),
		mu:          sync.Mutex{},
		config:      config,
		desktop:     desktop,
		streaming:   false,
		broadcast:   broadcastNew(config),
		screencast:  screencastNew(config),
		audio:       streamNew(config.AudioCodec, config.Device, config.AudioParams),
		video:       streamNew(config.VideoCodec, config.Display, config.VideoParams),
	}
}

func (manager *CaptureManagerCtx) Start() {
	if manager.broadcast.Enabled() {
		if err := manager.broadcast.createPipeline(); err != nil {
			manager.logger.Panic().Err(err).Msg("unable to create broadcast pipeline")
		}
	}

	manager.desktop.OnBeforeScreenSizeChange(func() {
		if manager.video.Enabled() {
			manager.video.destroyPipeline()
		}

		if manager.broadcast.Enabled() {
			manager.broadcast.destroyPipeline()
		}

		if manager.screencast.Started() {
			manager.screencast.destroyPipeline()
		}
	})

	manager.desktop.OnAfterScreenSizeChange(func() {
		if manager.video.Enabled() {
			if err := manager.video.createPipeline(); err != nil {
				manager.logger.Panic().Err(err).Msg("unable to recreate video pipeline")
			}
		}

		if manager.broadcast.Enabled() {
			if err := manager.broadcast.createPipeline(); err != nil {
				manager.logger.Panic().Err(err).Msg("unable to recreate broadcast pipeline")
			}
		}

		if manager.screencast.Started() {
			if err := manager.screencast.createPipeline(); err != nil {
				manager.logger.Panic().Err(err).Msg("unable to recreate screencast pipeline")
			}
		}
	})
}

func (manager *CaptureManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("capture shutting down")

	manager.broadcast.shutdown()
	manager.screencast.shutdown()

	manager.audio.shutdown()
	manager.video.shutdown()
	return nil
}

func (manager *CaptureManagerCtx) Broadcast() types.BroadcastManager {
	return manager.broadcast
}

func (manager *CaptureManagerCtx) Screencast() types.ScreencastManager {
	return manager.screencast
}

func (manager *CaptureManagerCtx) Audio() types.StreamManager {
	return manager.audio
}

func (manager *CaptureManagerCtx) Video() types.StreamManager {
	return manager.video
}

func (manager *CaptureManagerCtx) StartStream() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.logger.Info().Msgf("starting stream pipelines")

	var err error
	err = manager.Video().Start()
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to start video pipeline")
	}

	err = manager.Audio().Start()
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to start audio pipeline")
	}

	manager.streaming = true
}

func (manager *CaptureManagerCtx) StopStream() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.logger.Info().Msgf("stopping stream pipelines")

	manager.Video().Stop()
	manager.Audio().Stop()
	manager.streaming = false
}

func (manager *CaptureManagerCtx) Streaming() bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.streaming
}
