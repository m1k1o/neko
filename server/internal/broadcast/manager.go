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
	remote   *config.Remote
	config   *config.Broadcast
	enabled  bool
	url      string
}

func New(remote *config.Remote, config *config.Broadcast) *BroadcastManager {
	return &BroadcastManager{
		logger:  log.With().Str("module", "remote").Logger(),
		remote:  remote,
		config:  config,
		enabled: false,
		url:     "",
	}
}

func (manager *BroadcastManager) Start() {
	if !manager.enabled || manager.IsActive() {
		return
	}

	var err error
	manager.pipeline, err = gst.CreateRTMPPipeline(
		manager.remote.Device,
		manager.remote.Display,
		manager.config.Pipeline,
		manager.url,
	)

	manager.logger.Info().
		Str("audio_device", manager.remote.Device).
		Str("video_display", manager.remote.Display).
		Str("rtmp_pipeline_src", manager.pipeline.Src).
		Msgf("RTMP pipeline is starting...")

	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create rtmp pipeline")
		return
	}

	manager.pipeline.Play()
}

func (manager *BroadcastManager) Stop() {
	if !manager.IsActive() {
		return
	}

	manager.pipeline.Stop()
	manager.pipeline = nil
}

func (manager *BroadcastManager) IsActive() bool {
	return manager.pipeline != nil
}

func (manager *BroadcastManager) Create(url string) {
	manager.url = url
	manager.enabled = true
	manager.Start()
}

func (manager *BroadcastManager) Destroy() {
	manager.Stop()
	manager.enabled = false
}

func (manager *BroadcastManager) GetUrl() string {
	return manager.url
}
