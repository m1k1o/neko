package remote

import (
	"fmt"
	"time"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/gst"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/config"
	"demodesk/neko/internal/xorg"
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

	manager.logger.Info().
		Str("screen_resolution", fmt.Sprintf("%dx%d@%d", manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate)).
		Msgf("Setting screen resolution...")

	if !xorg.ValidScreenSize(manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate) {
		manager.logger.Warn().Msgf("invalid screen option %dx%d@%d", manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate)
	} else if err := xorg.ChangeScreenSize(manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate); err != nil {
		manager.logger.Warn().Err(err).Msg("unable to change screen size")
	}

	manager.createVideoPipeline()
	manager.createAudioPipeline()
	manager.broadcast.Start()

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
	manager.video.DestroyPipeline()
	manager.audio.DestroyPipeline()
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
	manager.logger.Info().Msgf("Pipelines starting...")

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

func (manager *RemoteManager) createVideoPipeline() {
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

func (manager *RemoteManager) createAudioPipeline() {
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

func (manager *RemoteManager) ChangeResolution(width int, height int, rate int) error {
	if !xorg.ValidScreenSize(width, height, rate) {
		return fmt.Errorf("unknown configuration")
	}

	manager.video.DestroyPipeline()
	manager.broadcast.Stop()

	defer func() {
		manager.video.Start()
		manager.broadcast.Start()

		manager.logger.Info().Msg("starting video pipeline...")
	}()

	if err := xorg.ChangeScreenSize(width, height, rate); err != nil {
		return err
	}

	manager.createVideoPipeline()
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

func (manager *RemoteManager) SetKeyboardModifiers(NumLock int, CapsLock int, ScrollLock int) {
	xorg.SetKeyboardModifiers(NumLock, CapsLock, ScrollLock)
}
