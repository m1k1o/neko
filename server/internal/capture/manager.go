package capture

import (
	"time"

	"m1k1o/neko/internal/capture/gst"
	"m1k1o/neko/internal/config"
	"m1k1o/neko/internal/desktop/xorg"
	"m1k1o/neko/internal/types"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type CaptureManagerCtx struct {
	logger    zerolog.Logger
	video     *gst.Pipeline
	audio     *gst.Pipeline
	config    *config.Capture
	broadcast types.BroadcastManager
	desktop   types.DesktopManager
	cleanup   *time.Ticker
	shutdown  chan bool
	emmiter   events.EventEmmiter
	streaming bool
}

func New(desktop types.DesktopManager, broadcast types.BroadcastManager, config *config.Capture) *CaptureManagerCtx {
	return &CaptureManagerCtx{
		logger:    log.With().Str("module", "capture").Logger(),
		cleanup:   time.NewTicker(1 * time.Second),
		shutdown:  make(chan bool),
		emmiter:   events.New(),
		config:    config,
		broadcast: broadcast,
		desktop:   desktop,
		streaming: false,
	}
}

func (manager *CaptureManagerCtx) VideoCodec() string {
	return manager.config.VideoCodec
}

func (manager *CaptureManagerCtx) AudioCodec() string {
	return manager.config.AudioCodec
}

func (manager *CaptureManagerCtx) Start() {
	manager.createPipelines()
	if err := manager.broadcast.Start(); err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create rtmp pipeline")
	}

	go func() {
		defer func() {
			manager.logger.Info().Msg("shutdown")
		}()

		for {
			select {
			case <-manager.shutdown:
				return
			case sample := <-manager.video.Sample:
				manager.emmiter.Emit("video", sample)
			case sample := <-manager.audio.Sample:
				manager.emmiter.Emit("audio", sample)
			case <-manager.cleanup.C:
				// TODO: refactor.
			}
		}
	}()
}

func (manager *CaptureManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("capture shutting down")
	manager.video.Stop()
	manager.audio.Stop()
	manager.broadcast.Stop()

	manager.cleanup.Stop()
	manager.shutdown <- true
	return nil
}

func (manager *CaptureManagerCtx) OnVideoFrame(listener func(sample types.Sample)) {
	manager.emmiter.On("video", func(payload ...interface{}) {
		listener(payload[0].(types.Sample))
	})
}

func (manager *CaptureManagerCtx) OnAudioFrame(listener func(sample types.Sample)) {
	manager.emmiter.On("audio", func(payload ...interface{}) {
		listener(payload[0].(types.Sample))
	})
}

func (manager *CaptureManagerCtx) StartStream() {
	manager.createPipelines()

	manager.logger.Info().
		Str("video_display", manager.config.Display).
		Str("video_codec", manager.config.VideoCodec).
		Str("audio_device", manager.config.Device).
		Str("audio_codec", manager.config.AudioCodec).
		Str("audio_pipeline_src", manager.audio.Src).
		Str("video_pipeline_src", manager.video.Src).
		Msgf("Pipelines starting...")

	manager.video.Start()
	manager.audio.Start()
	manager.streaming = true
}

func (manager *CaptureManagerCtx) StopStream() {
	manager.logger.Info().Msgf("Pipelines shutting down...")
	manager.video.Stop()
	manager.audio.Stop()
	manager.streaming = false
}

func (manager *CaptureManagerCtx) Streaming() bool {
	return manager.streaming
}

func (manager *CaptureManagerCtx) createPipelines() {
	// handle maximum fps
	rate := manager.desktop.GetScreenSize().Rate
	if manager.config.MaxFPS != 0 && manager.config.MaxFPS < rate {
		rate = manager.config.MaxFPS
	}

	var err error
	manager.video, err = CreateAppPipeline(
		manager.config.VideoCodec,
		manager.config.Display,
		manager.config.VideoParams,
		rate,
		manager.config.VideoBitrate,
		manager.config.VideoHWEnc,
	)
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create video pipeline")
	}

	manager.audio, err = CreateAppPipeline(
		manager.config.AudioCodec,
		manager.config.Device,
		manager.config.AudioParams,
		0, // fps: n/a for audio
		manager.config.AudioBitrate,
		"", // hwenc: n/a for audio
	)
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create audio pipeline")
	}
}

func (manager *CaptureManagerCtx) ChangeResolution(width int, height int, rate int16) error {
	manager.video.Stop()
	manager.broadcast.Stop()

	defer func() {
		manager.video.Start()
		if err := manager.broadcast.Start(); err != nil {
			manager.logger.Panic().Err(err).Msg("unable to create rtmp pipeline")
		}

		manager.logger.Info().Msg("starting video pipeline...")
	}()

	if err := xorg.ChangeScreenSize(width, height, rate); err != nil {
		return err
	}

	// handle maximum fps
	if manager.config.MaxFPS != 0 && manager.config.MaxFPS < rate {
		rate = manager.config.MaxFPS
	}

	var err error
	manager.video, err = CreateAppPipeline(
		manager.config.VideoCodec,
		manager.config.Display,
		manager.config.VideoParams,
		rate,
		manager.config.VideoBitrate,
		manager.config.VideoHWEnc,
	)
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create new video pipeline")
	}

	return nil
}
