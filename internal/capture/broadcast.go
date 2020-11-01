package capture

import (
	"demodesk/neko/internal/capture/gst"
)

func (manager *CaptureManagerCtx) StartBroadcastPipeline() {
	var err error

	if manager.IsBoradcasting() || manager.broadcast_url == "" {
		return
	}

	manager.logger.Info().
		Str("audio_device", manager.config.Device).
		Str("video_display", manager.config.Display).
		Str("rtmp_pipeline_src", manager.broadcast.Src).
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
	if !manager.IsBoradcasting() {
		return
	}

	manager.broadcast.DestroyPipeline()
	manager.broadcast = nil
}

func (manager *CaptureManagerCtx) StartBroadcast(url string) {
	manager.broadcast_url = url
	manager.StartBroadcastPipeline()
}

func (manager *CaptureManagerCtx) StopBroadcast() {
	manager.broadcast_url = ""
	manager.StopBroadcastPipeline()
}

func (manager *CaptureManagerCtx) IsBoradcasting() bool {
	return manager.broadcast != nil
}

func (manager *CaptureManagerCtx) BroadcastUrl() string {
	return manager.broadcast_url
}
