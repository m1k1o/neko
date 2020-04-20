package broadcast

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"n.eko.moe/neko/internal/gst"
	"n.eko.moe/neko/internal/types/config"
)

type BroadcastManager struct {
	logger   zerolog.Logger
	pipeline *gst.Pipeline
	config   *config.Broadcast
}

func New(config *config.Broadcast) *BroadcastManager {
	return &BroadcastManager{
		logger: log.With().Str("module", "remote").Logger(),
		config: config,
	}
}

func (manager *BroadcastManager) Start() {
	var err error
	manager.pipeline, err = gst.CreateRTMPPipeline(
		manager.config.Device,
		manager.config.Display,
		manager.config.RTMP,
	)
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create rtmp pipeline")
	}

	manager.pipeline.Start()
}

func (manager *BroadcastManager) Shutdown() error {
	if manager.pipeline != nil {
		manager.pipeline.Stop()
	}

	return nil
}
