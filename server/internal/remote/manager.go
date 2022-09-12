package remote

import (
	"fmt"
	"time"

	"m1k1o/neko/internal/gst"
	"m1k1o/neko/internal/remote/xorg"
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/types/config"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type RemoteManager struct {
	logger    zerolog.Logger
	video     *gst.Pipeline
	audio     *gst.Pipeline
	config    *config.Remote
	broadcast types.BroadcastManager
	cleanup   *time.Ticker
	shutdown  chan bool
	emmiter   events.EventEmmiter
	streaming bool
}

func New(config *config.Remote, broadcast types.BroadcastManager) *RemoteManager {
	return &RemoteManager{
		logger:    log.With().Str("module", "remote").Logger(),
		cleanup:   time.NewTicker(1 * time.Second),
		shutdown:  make(chan bool),
		emmiter:   events.New(),
		config:    config,
		broadcast: broadcast,
		streaming: false,
	}
}

func (manager *RemoteManager) VideoCodec() string {
	return manager.config.VideoCodec
}

func (manager *RemoteManager) AudioCodec() string {
	return manager.config.AudioCodec
}

func (manager *RemoteManager) Start() {
	xorg.Display(manager.config.Display)

	if !xorg.ValidScreenSize(manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate) {
		manager.logger.Warn().Msgf("invalid screen option %dx%d@%d", manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate)
	} else if err := xorg.ChangeScreenSize(manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate); err != nil {
		manager.logger.Warn().Err(err).Msg("unable to change screen size")
	}

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
				xorg.CheckKeys(time.Second * 10)
			}
		}
	}()
}

func (manager *RemoteManager) Shutdown() error {
	manager.logger.Info().Msgf("remote shutting down")
	manager.video.Stop()
	manager.audio.Stop()
	manager.broadcast.Stop()

	manager.cleanup.Stop()
	manager.shutdown <- true
	return nil
}

func (manager *RemoteManager) OnVideoFrame(listener func(sample types.Sample)) {
	manager.emmiter.On("video", func(payload ...interface{}) {
		listener(payload[0].(types.Sample))
	})
}

func (manager *RemoteManager) OnAudioFrame(listener func(sample types.Sample)) {
	manager.emmiter.On("audio", func(payload ...interface{}) {
		listener(payload[0].(types.Sample))
	})
}

func (manager *RemoteManager) StartStream() {
	manager.createPipelines()

	manager.logger.Info().
		Str("video_display", manager.config.Display).
		Str("video_codec", manager.config.VideoCodec).
		Str("audio_device", manager.config.Device).
		Str("audio_codec", manager.config.AudioCodec).
		Str("audio_pipeline_src", manager.audio.Src).
		Str("video_pipeline_src", manager.video.Src).
		Str("screen_resolution", fmt.Sprintf("%dx%d@%d", manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate)).
		Msgf("Pipelines starting...")

	manager.video.Start()
	manager.audio.Start()
	manager.streaming = true
}

func (manager *RemoteManager) StopStream() {
	manager.logger.Info().Msgf("Pipelines shutting down...")
	manager.video.Stop()
	manager.audio.Stop()
	manager.streaming = false
}

func (manager *RemoteManager) Streaming() bool {
	return manager.streaming
}

func (manager *RemoteManager) createPipelines() {
	// handle maximum fps
	rate := manager.config.ScreenRate
	if manager.config.MaxFPS != 0 && manager.config.MaxFPS < manager.config.ScreenRate {
		rate = manager.config.MaxFPS
	}

	var err error
	manager.video, err = gst.CreateAppPipeline(
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

	manager.audio, err = gst.CreateAppPipeline(
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

func (manager *RemoteManager) ChangeResolution(width int, height int, rate int) error {
	if !xorg.ValidScreenSize(width, height, rate) {
		return fmt.Errorf("unknown configuration")
	}

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
	manager.video, err = gst.CreateAppPipeline(
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
