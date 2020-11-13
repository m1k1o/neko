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
	emit_update     chan bool
	emit_stop       chan bool
	video_sample    chan types.Sample
	audio_sample    chan types.Sample
	emmiter         events.EventEmmiter
	streaming       bool
	broadcasting    bool
	broadcast_url   string
	desktop         types.DesktopManager
}

func New(desktop types.DesktopManager, config *config.Capture) *CaptureManagerCtx {
	return &CaptureManagerCtx{
		logger:          log.With().Str("module", "capture").Logger(),
		emit_update:     make(chan bool),
		emit_stop:       make(chan bool),
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
		manager.destroyVideoPipeline()
		manager.StopBroadcastPipeline()
	})

	manager.desktop.OnAfterScreenSizeChange(func() {
		manager.createVideoPipeline()
		manager.StartBroadcastPipeline()
	})

	go func() {
		manager.logger.Debug().Msg("started emitting samples")

		for {
			select {
			case <-manager.emit_stop:
				manager.logger.Debug().Msg("stopped emitting samples")
				return
			case <-manager.emit_update:
				manager.logger.Debug().Msg("update emitting samples")
			case sample := <-manager.video_sample:
				manager.emmiter.Emit("video", sample)
			case sample := <-manager.audio_sample:
				manager.emmiter.Emit("audio", sample)
			}
		}
	}()
}

func (manager *CaptureManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("capture shutting down")
	manager.emit_stop <- true
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

	manager.destroyVideoPipeline()
	manager.destroyAudioPipeline()
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

	manager.video_sample = manager.video.Sample
	manager.emit_update <-true
}

func (manager *CaptureManagerCtx) destroyVideoPipeline() {
	manager.logger.Info().Msgf("stopping video pipeline")
	manager.video.Stop()
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

	manager.audio_sample = manager.audio.Sample
	manager.emit_update <-true
}

func (manager *CaptureManagerCtx) destroyAudioPipeline() {
	manager.logger.Info().Msgf("stopping audio pipeline")
	manager.audio.Stop()
}
