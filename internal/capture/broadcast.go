package capture

import (
	"demodesk/neko/internal/capture/gst"
)

func (manager *CaptureManagerCtx) StartBroadcastPipeline() {
	var err error

	if manager.broadcast != nil || !manager.BroadcastEnabled() {
		return
	}

	manager.logger.Info().
		Str("audio_device", manager.config.Device).
		Str("video_display", manager.config.Display).
		Str("broadcast_pipeline", manager.config.BroadcastPipeline).
		Msgf("Creating broadcast pipeline...")
	
	manager.broadcast, err = gst.CreateRTMPPipeline(
		manager.config.Device,
		manager.config.Display,
		manager.config.BroadcastPipeline,
		manager.broadcast_url,
	)

	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create broadcast pipeline")
	}

	manager.broadcast.Play()
	manager.logger.Info().Msgf("Starting broadcast pipeline...")
}

func (manager *CaptureManagerCtx) StopBroadcastPipeline() {
	if manager.broadcast == nil {
		return
	}

	manager.broadcast.DestroyPipeline()
	manager.logger.Info().Msgf("Stopping broadcast pipeline...")
	manager.broadcast = nil
}

func (manager *CaptureManagerCtx) StartBroadcast(url string) {
	manager.broadcast_url = url
	manager.broadcasting = true
	manager.StartBroadcastPipeline()
}

func (manager *CaptureManagerCtx) StopBroadcast() {
	manager.broadcasting = false
	manager.StopBroadcastPipeline()
}

func (manager *CaptureManagerCtx) BroadcastEnabled() bool {
	return manager.broadcasting
}

func (manager *CaptureManagerCtx) BroadcastUrl() string {
	return manager.broadcast_url
}
