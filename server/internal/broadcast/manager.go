package broadcast

import (
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/gst"
	"m1k1o/neko/internal/types/config"
)

type BroadcastManager struct {
	mu       sync.Mutex
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

func (manager *BroadcastManager) Shutdown() error {
	manager.Destroy()
	return nil
}

func (manager *BroadcastManager) Start() error {
	if !manager.enabled || manager.IsActive() {
		return nil
	}

	var err error
	manager.pipeline, err = gst.CreateRTMPPipeline(
		manager.remote.Device,
		manager.remote.Display,
		manager.config.Pipeline,
		manager.url,
	)

	if err != nil {
		manager.pipeline = nil
		return err
	}

	manager.logger.Info().
		Str("audio_device", manager.remote.Device).
		Str("video_display", manager.remote.Display).
		Str("rtmp_pipeline_src", manager.pipeline.Src).
		Msgf("RTMP pipeline is starting...")

	manager.pipeline.Play()
	return nil
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

func (manager *BroadcastManager) Create(url string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.url = url
	manager.enabled = true

	err := manager.Start()
	if err != nil {
		manager.enabled = false
	}

	return err
}

func (manager *BroadcastManager) Destroy() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.Stop()
	manager.enabled = false
}

func (manager *BroadcastManager) GetUrl() string {
	return manager.url
}
