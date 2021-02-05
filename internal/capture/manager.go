package capture

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/config"
)

type CaptureManagerCtx struct {
	logger       zerolog.Logger
	mu           sync.Mutex
	desktop      types.DesktopManager
	streaming    bool
	broadcast    *BroacastManagerCtx
	screencast   *ScreencastManagerCtx
	audio        *StreamManagerCtx
	videos       map[string]*StreamManagerCtx
	videoIDs     []string
}

func New(desktop types.DesktopManager, config *config.Capture) *CaptureManagerCtx {
	logger := log.With().Str("module", "capture").Logger()

	broadcastPipeline := config.BroadcastPipeline
	if broadcastPipeline == "" {
		broadcastPipeline = fmt.Sprintf(
			"flvmux name=mux ! rtmpsink location='{url} live=1' " +
				"pulsesrc device=%s " +
					"! audio/x-raw,channels=2 " +
					"! audioconvert " +
					"! queue " +
					"! voaacenc " +
					"! mux. " +
				"ximagesrc display-name=%s show-pointer=true use-damage=false " +
					"! video/x-raw " +
					"! videoconvert " +
					"! queue " +
					"! x264enc threads=4 bitrate=4096 key-int-max=15 byte-stream=true byte-stream=true tune=zerolatency speed-preset=veryfast " +
					"! mux.", config.Device, config.Display,
		)
	}

	screencastPipeline := config.ScreencastPipeline
	if screencastPipeline == "" {
		screencastPipeline = fmt.Sprintf(
			"ximagesrc display-name=%s show-pointer=true use-damage=false " +
				"! video/x-raw,framerate=%s " +
				"! videoconvert " +
				"! queue " +
				"! jpegenc quality=%s " +
				"! appsink name=appsink", config.Display, config.ScreencastRate, config.ScreencastQuality,
		)
	}

	audioPipeline := config.AudioPipeline
	if audioPipeline == "" {
		audioPipeline = fmt.Sprintf(
			"pulsesrc device=%s " +
				"! audio/x-raw,channels=2 " +
				"! audioconvert " +
				"! queue " +
				"! %s " +
				"! appsink name=appsink", config.Device, config.AudioCodec.Pipeline,
		)
	}

	videoPipeline := config.VideoPipeline
	if videoPipeline == "" {
		videoPipeline = fmt.Sprintf(
			"ximagesrc display-name=%s show-pointer=false use-damage=false " +
				"! video/x-raw " +
				"! videoconvert " +
				"! queue " +
				"! %s " +
				"! appsink name=appsink", config.Display, config.VideoCodec.Pipeline,
		)
	}

	return &CaptureManagerCtx{
		logger:      logger,
		desktop:     desktop,
		streaming:   false,
		broadcast:   broadcastNew(broadcastPipeline),
		screencast:  screencastNew(config.Screencast, screencastPipeline),
		audio:       streamNew(config.AudioCodec, audioPipeline),
		videos:      map[string]*StreamManagerCtx{
			"hq": streamNew(config.VideoCodec, videoPipeline),
		},
		videoIDs:    []string{ "hq" },
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

func (manager *CaptureManagerCtx) StartStream() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.logger.Info().Msgf("starting stream pipelines")

	for _, video := range manager.videos {
		if err := video.Start(); err != nil {
			manager.logger.Panic().Err(err).Msg("unable to start video pipeline")
		}
	}

	if err := manager.audio.Start(); err != nil {
		manager.logger.Panic().Err(err).Msg("unable to start audio pipeline")
	}

	manager.streaming = true
}

func (manager *CaptureManagerCtx) StopStream() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.logger.Info().Msgf("stopping stream pipelines")

	for _, video := range manager.videos {
		video.Stop()
	}

	manager.audio.Stop()
	manager.streaming = false
}

func (manager *CaptureManagerCtx) Streaming() bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	return manager.streaming
}
