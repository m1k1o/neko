package capture

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/types"
)

type CaptureManagerCtx struct {
	logger     zerolog.Logger
	desktop    types.DesktopManager
	streaming  bool
	broadcast  *BroacastManagerCtx
	screencast *ScreencastManagerCtx
	audio      *StreamManagerCtx
	videos     map[string]*StreamManagerCtx
	videoIDs   []string
}

func New(desktop types.DesktopManager, config *config.Capture) *CaptureManagerCtx {
	logger := log.With().Str("module", "capture").Logger()

	broadcastPipeline := config.BroadcastPipeline
	if broadcastPipeline == "" {
		broadcastPipeline = fmt.Sprintf(
			"flvmux name=mux ! rtmpsink location='{url} live=1' "+
				"pulsesrc device=%s "+
				"! audio/x-raw,channels=2 "+
				"! audioconvert "+
				"! queue "+
				"! voaacenc bitrate=%d "+
				"! mux. "+
				"ximagesrc display-name=%s show-pointer=true use-damage=false "+
				"! video/x-raw "+
				"! videoconvert "+
				"! queue "+
				"! x264enc threads=4 bitrate=%d key-int-max=15 byte-stream=true tune=zerolatency speed-preset=%s "+
				"! mux.", config.AudioDevice, config.BroadcastAudioBitrate*1000, config.Display, config.BroadcastVideoBitrate, config.BroadcastPreset,
		)
	}

	screencastPipeline := config.ScreencastPipeline
	if screencastPipeline == "" {
		screencastPipeline = fmt.Sprintf(
			"ximagesrc display-name=%s show-pointer=true use-damage=false "+
				"! video/x-raw,framerate=%s "+
				"! videoconvert "+
				"! queue "+
				"! jpegenc quality=%s "+
				"! appsink name=appsink", config.Display, config.ScreencastRate, config.ScreencastQuality,
		)
	}

	videos := map[string]*StreamManagerCtx{}
	videoIDs := []string{}
	for key, pipelineConf := range config.Video {
		codec, err := pipelineConf.GetCodec()
		if err != nil {
			logger.Panic().Err(err).Str("video_key", key).Msg("unable to get video codec")
		}

		createPipeline := func() string {
			screen := desktop.GetScreenSize()

			pipeline, err := pipelineConf.GetPipeline(*screen)
			if err != nil {
				logger.Panic().Err(err).Str("video_key", key).Msg("unable to get video pipeline")
			}

			return fmt.Sprintf(
				"ximagesrc display-name=%s show-pointer=false use-damage=false "+
					"! %s "+
					"! appsink name=appsink", config.Display, pipeline,
			)
		}

		// trigger function to catch evaluation errors at startup
		_ = createPipeline()

		// append to videos
		videos[key] = streamNew(codec, createPipeline)
		videoIDs = append(videoIDs, key)
	}

	return &CaptureManagerCtx{
		logger:     logger,
		desktop:    desktop,
		streaming:  false,
		broadcast:  broadcastNew(broadcastPipeline),
		screencast: screencastNew(config.ScreencastEnabled, screencastPipeline),
		audio: streamNew(config.AudioCodec, func() string {
			if config.AudioPipeline != "" {
				return config.AudioPipeline
			}

			return fmt.Sprintf(
				"pulsesrc device=%s "+
					"! audio/x-raw,channels=2 "+
					"! audioconvert "+
					"! queue "+
					"! %s "+
					"! appsink name=appsink", config.AudioDevice, config.AudioCodec.Pipeline,
			)
		}),
		videos:   videos,
		videoIDs: videoIDs,
	}
}

func (manager *CaptureManagerCtx) Start() {
	if manager.broadcast.Started() {
		if err := manager.broadcast.createPipeline(); err != nil {
			manager.logger.Panic().Err(err).Msg("unable to create broadcast pipeline")
		}
	}

	manager.desktop.OnBeforeScreenSizeChange(func() {
		for _, video := range manager.videos {
			if video.Started() {
				video.destroyPipeline()
			}
		}

		if manager.broadcast.Started() {
			manager.broadcast.destroyPipeline()
		}

		if manager.screencast.Started() {
			manager.screencast.destroyPipeline()
		}
	})

	manager.desktop.OnAfterScreenSizeChange(func() {
		for _, video := range manager.videos {
			if video.Started() {
				if err := video.createPipeline(); err != nil {
					manager.logger.Panic().Err(err).Msg("unable to recreate video pipeline")
				}
			}
		}

		if manager.broadcast.Started() {
			if err := manager.broadcast.createPipeline(); err != nil {
				manager.logger.Panic().Err(err).Msg("unable to recreate broadcast pipeline")
			}
		}

		if manager.screencast.Started() {
			if err := manager.screencast.createPipeline(); err != nil {
				manager.logger.Panic().Err(err).Msg("unable to recreate screencast pipeline")
			}
		}
	})
}

func (manager *CaptureManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("capture shutting down")

	manager.broadcast.shutdown()
	manager.screencast.shutdown()

	manager.audio.shutdown()

	for _, video := range manager.videos {
		video.shutdown()
	}

	return nil
}

func (manager *CaptureManagerCtx) Broadcast() types.BroadcastManager {
	return manager.broadcast
}

func (manager *CaptureManagerCtx) Screencast() types.ScreencastManager {
	return manager.screencast
}

func (manager *CaptureManagerCtx) Audio() types.StreamManager {
	return manager.audio
}

func (manager *CaptureManagerCtx) Video(videoID string) (types.StreamManager, bool) {
	video, ok := manager.videos[videoID]
	return video, ok
}

func (manager *CaptureManagerCtx) VideoIDs() []string {
	return manager.videoIDs
}
