package remote

import (
	"fmt"
	"time"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"n.eko.moe/neko/internal/gst"
	"n.eko.moe/neko/internal/types"
	"n.eko.moe/neko/internal/types/config"
	"n.eko.moe/neko/internal/xorg"
)

type RemoteManager struct {
	logger    zerolog.Logger
	video     *gst.Pipeline
	audio     *gst.Pipeline
	config    *config.Remote
	cleanup   *time.Ticker
	shutdown  chan bool
	emmiter   events.EventEmmiter
	streaming bool
}

func New(config *config.Remote) *RemoteManager {
	return &RemoteManager{
		logger:    log.With().Str("module", "remote").Logger(),
		cleanup:   time.NewTicker(1 * time.Second),
		shutdown:  make(chan bool),
		emmiter:   events.New(),
		config:    config,
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
	manager.createPipelines()

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
	manager.logger.Info().
		Str("video_display", manager.config.Display).
		Str("video_codec", manager.config.VideoCodec).
		Str("audio_device", manager.config.Device).
		Str("audio_codec", manager.config.AudioCodec).
		Str("audio_pipeline_src", manager.audio.Src).
		Str("video_pipeline_src", manager.video.Src).
		Str("screen_resolution", fmt.Sprintf("%dx%d@%d", manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate)).
		Msgf("Pipelines starting...")

	xorg.Display(manager.config.Display)

	if !xorg.ValidScreenSize(manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate) {
		manager.logger.Warn().Msgf("invalid screen option %dx%d@%d", manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate)
	} else if err := xorg.ChangeScreenSize(manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate); err != nil {
		manager.logger.Warn().Err(err).Msg("unable to change screen size")
	}

	manager.createPipelines()
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
	var err error
	manager.video, err = gst.CreateAppPipeline(
		manager.config.VideoCodec,
		manager.config.Display,
		manager.config.VideoParams,
	)
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create video pipeline")
	}

	manager.audio, err = gst.CreateAppPipeline(
		manager.config.AudioCodec,
		manager.config.Device,
		manager.config.AudioParams,
	)
	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to screate audio pipeline")
	}
}

func (manager *RemoteManager) ChangeResolution(width int, height int, rate int) error {
	if !xorg.ValidScreenSize(width, height, rate) {
		return fmt.Errorf("unknown configuration")
	}

	manager.video.Stop()
	defer func() {
		manager.video.Start()
		manager.logger.Info().Msg("starting video pipeline...")
	}()

	if err := xorg.ChangeScreenSize(width, height, rate); err != nil {
		return err
	}

	video, err := gst.CreateAppPipeline(
		manager.config.VideoCodec,
		manager.config.Display,
		manager.config.VideoParams,
	)

	if err != nil {
		manager.logger.Panic().Err(err).Msg("unable to create new video pipeline")
	}

	manager.video = video
	return nil
}

func (manager *RemoteManager) Move(x, y int) {
	xorg.Move(x, y)
}

func (manager *RemoteManager) Scroll(x, y int) {
	xorg.Scroll(x, y)
}

func (manager *RemoteManager) ButtonDown(code int) error {
	return xorg.ButtonDown(code)
}

func (manager *RemoteManager) KeyDown(code uint64) error {
	return xorg.KeyDown(code)
}

func (manager *RemoteManager) ButtonUp(code int) error {
	return xorg.ButtonUp(code)
}

func (manager *RemoteManager) KeyUp(code uint64) error {
	return xorg.KeyUp(code)
}

func (manager *RemoteManager) ReadClipboard() string {
	return xorg.ReadClipboard()
}

func (manager *RemoteManager) WriteClipboard(data string) {
	xorg.WriteClipboard(data)
}

func (manager *RemoteManager) ResetKeys() {
	xorg.ResetKeys()
}

func (manager *RemoteManager) ScreenConfigurations() map[int]types.ScreenConfiguration {
	return xorg.ScreenConfigurations
}

func (manager *RemoteManager) GetScreenSize() *types.ScreenSize {
	return xorg.GetScreenSize()
}

func (manager *RemoteManager) SetKeyboardLayout(layout string) {
	xorg.SetKeyboardLayout(layout)
}