package capture

import (
	"fmt"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/config"
	"demodesk/neko/internal/capture/gst"
)

type CaptureManagerCtx struct {
	logger        zerolog.Logger
	video         *gst.Pipeline
	audio         *gst.Pipeline
	broadcast     *gst.Pipeline
	config        *config.Capture
	shutdown      chan bool
	emmiter       events.EventEmmiter
	streaming     bool
	broadcasting  bool
	broadcast_url string
	desktop       types.DesktopManager
}

func New(desktop types.DesktopManager, config *config.Capture) *CaptureManagerCtx {
	return &CaptureManagerCtx{
		logger:        log.With().Str("module", "capture").Logger(),
		shutdown:      make(chan bool),
		emmiter:       events.New(),
		config:        config,
		streaming:     false,
		broadcasting:  false,
		broadcast_url: "",
		desktop:       desktop,
	}
}

func (manager *CaptureManagerCtx) Start() {
	manager.logger.Info().
		Str("screen_resolution", fmt.Sprintf("%dx%d@%d", manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate)).
		Msgf("Setting screen resolution...")

	if err := manager.desktop.ChangeScreenSize(manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate); err != nil {
		manager.logger.Warn().Err(err).Msg("unable to change screen size")
	}

	manager.CreateVideoPipeline()
	manager.CreateAudioPipeline()
	manager.StartBroadcastPipeline()

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
			}
		}
	}()
}

func (manager *CaptureManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("capture shutting down")
	manager.video.DestroyPipeline()
	manager.audio.DestroyPipeline()
	manager.StopBroadcastPipeline()

	manager.shutdown <- true
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
	manager.logger.Info().Msgf("Pipelines starting...")

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

func (manager *CaptureManagerCtx) CreateVideoPipeline() {
	var err error

	manager.logger.Info().
		Str("video_codec", manager.config.VideoCodec).
		Str("video_display", manager.config.Display).
		Str("video_params", manager.config.VideoParams).
		Msgf("Creating video pipeline...")

	manager.video, err = gst.CreateAppPipeline(
		manager.config.VideoCodec,
		manager.config.Display,
		manager.config.VideoParams,
	)

	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create video pipeline")
	}
}

func (manager *CaptureManagerCtx) CreateAudioPipeline() {
	var err error

	manager.logger.Info().
		Str("audio_codec", manager.config.AudioCodec).
		Str("audio_display", manager.config.Device).
		Str("audio_params", manager.config.AudioParams).
		Msgf("Creating audio pipeline...")

	manager.audio, err = gst.CreateAppPipeline(
		manager.config.AudioCodec,
		manager.config.Device,
		manager.config.AudioParams,
	)

	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create audio pipeline")
	}
}

func (manager *CaptureManagerCtx) ChangeResolution(width int, height int, rate int) error {
	manager.video.DestroyPipeline()
	manager.StopBroadcastPipeline()

	defer func() {
		manager.CreateVideoPipeline()
	
		manager.video.Start()
		manager.logger.Info().Msg("starting video pipeline...")
	
		manager.StartBroadcastPipeline()
	}()
	
	return manager.desktop.ChangeScreenSize(width, height, rate)
}
