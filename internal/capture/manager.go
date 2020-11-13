package capture

import (
	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/config"
	"demodesk/neko/internal/capture/gst"
)

type CaptureManagerCtx struct {
	logger          zerolog.Logger
	video           *gst.Pipeline
	audio           *gst.Pipeline
	broadcast       *gst.Pipeline
	config          *config.Capture
	audio_emit_stop chan bool
	video_emit_stop chan bool
	emmiter         events.EventEmmiter
	streaming       bool
	broadcasting    bool
	broadcast_url   string
	desktop         types.DesktopManager
}

func New(desktop types.DesktopManager, config *config.Capture) *CaptureManagerCtx {
	return &CaptureManagerCtx{
		logger:          log.With().Str("module", "capture").Logger(),
		audio_emit_stop: make(chan bool),
		video_emit_stop: make(chan bool),
		emmiter:         events.New(),
		config:          config,
		streaming:       false,
		broadcasting:    false,
		broadcast_url:   "",
		desktop:         desktop,
	}
}

func (manager *CaptureManagerCtx) Start() {
	manager.StartBroadcastPipeline()

	manager.desktop.OnBeforeScreenSizeChange(func() {
		manager.video_emit_stop <- true
		manager.logger.Info().Msgf("stopping video pipeline")
		manager.video.Stop()

		manager.StopBroadcastPipeline()
	})

	manager.desktop.OnAfterScreenSizeChange(func() {
		manager.createVideoPipeline()
		manager.StartBroadcastPipeline()
	})
}

func (manager *CaptureManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("capture shutting down")
	manager.StopStream()
	return nil
}

func (manager *CaptureManagerCtx) VideoCodec() string {
	return manager.config.VideoCodec
}

func (manager *CaptureManagerCtx) AudioCodec() string {
	return manager.config.AudioCodec
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
	manager.logger.Info().Msgf("starting pipelines")

	manager.createVideoPipeline()
	manager.createAudioPipeline()
	manager.streaming = true
}

func (manager *CaptureManagerCtx) StopStream() {
	manager.logger.Info().Msgf("stopping pipelines")

	manager.audio_emit_stop <- true
	manager.logger.Info().Msgf("stopping video pipeline")
	manager.audio.Stop()

	manager.video_emit_stop <- true
	manager.logger.Info().Msgf("stopping audio pipeline")
	manager.video.Stop()

	manager.streaming = false
}

func (manager *CaptureManagerCtx) Streaming() bool {
	return manager.streaming
}

func (manager *CaptureManagerCtx) createVideoPipeline() {
	var err error

	manager.logger.Info().
		Str("video_codec", manager.config.VideoCodec).
		Str("video_display", manager.config.Display).
		Str("video_params", manager.config.VideoParams).
		Msgf("creating video pipeline")

	manager.video, err = gst.CreateAppPipeline(
		manager.config.VideoCodec,
		manager.config.Display,
		manager.config.VideoParams,
	)

	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create video pipeline")
	}

	manager.logger.Info().
		Str("src", manager.video.Src).
		Msgf("starting video pipeline")

	manager.video.Start()

	go func() {
		manager.logger.Debug().Msg("started emitting video samples")

		for {
			select {
			case <-manager.video_emit_stop:
				manager.logger.Debug().Msg("stopped emitting video samples")
				return
			case sample := <-manager.video.Sample:
				manager.emmiter.Emit("video", sample)
			}
		}
	}()
}

func (manager *CaptureManagerCtx) createAudioPipeline() {
	var err error

	manager.logger.Info().
		Str("audio_codec", manager.config.AudioCodec).
		Str("audio_display", manager.config.Device).
		Str("audio_params", manager.config.AudioParams).
		Msgf("creating audio pipeline")

	manager.audio, err = gst.CreateAppPipeline(
		manager.config.AudioCodec,
		manager.config.Device,
		manager.config.AudioParams,
	)

	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create audio pipeline")
	}

	manager.logger.Info().
		Str("src", manager.audio.Src).
		Msgf("starting audio pipeline")

	manager.audio.Start()

	go func() {
		manager.logger.Debug().Msg("started emitting audio samples")

		for {
			select {
			case <-manager.audio_emit_stop:
				manager.logger.Debug().Msg("stopped emitting audio samples")
				return
			case sample := <-manager.audio.Sample:
				manager.emmiter.Emit("audio", sample)
			}
		}
	}()
}
