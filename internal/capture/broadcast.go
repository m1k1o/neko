package capture

import (
	"demodesk/neko/internal/capture/gst"
)

func (manager *CaptureManagerCtx) StartBroadcast(url string) error {
	manager.broadcast_url = url
	manager.broadcasting = true
	return manager.createBroadcastPipeline()
}

func (manager *CaptureManagerCtx) StopBroadcast() {
	manager.broadcasting = false
	manager.destroyBroadcastPipeline()
}

func (manager *CaptureManagerCtx) BroadcastEnabled() bool {
	return manager.broadcasting
}

func (manager *CaptureManagerCtx) BroadcastUrl() string {
	return manager.broadcast_url
}

func (manager *CaptureManagerCtx) createBroadcastPipeline() error {
	var err error

	manager.logger.Info().
		Str("audio_device", manager.config.Device).
		Str("video_display", manager.config.Display).
		Str("broadcast_pipeline", manager.config.BroadcastPipeline).
		Msgf("creating broadcast pipeline")
	
	manager.broadcast, err = gst.CreateRTMPPipeline(
		manager.config.Device,
		manager.config.Display,
		manager.config.BroadcastPipeline,
		manager.broadcast_url,
	)

	if err != nil {
		return err
	}

	manager.broadcast.Play()
	manager.logger.Info().Msgf("starting broadcast pipeline")
	return nil
}

func (manager *CaptureManagerCtx) destroyBroadcastPipeline() {
	if manager.broadcast == nil {
		return
	}

	manager.broadcast.Stop()
	manager.logger.Info().Msgf("stopping broadcast pipeline")
	manager.broadcast = nil
}
