package capture

import (
	"fmt"
	"math"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/codec"
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
				"! voaacenc bitrate=128000 "+
				"! mux. "+
				"ximagesrc display-name=%s show-pointer=true use-damage=false "+
				"! video/x-raw "+
				"! videoconvert "+
				"! queue "+
				"! x264enc threads=4 bitrate=4096 key-int-max=15 byte-stream=true tune=zerolatency speed-preset=veryfast "+
				"! mux.", config.AudioDevice, config.Display,
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
		videos: map[string]*StreamManagerCtx{
			"hd": streamNew(codec.VP8(), func() string {
				screen := desktop.GetScreenSize()
				bitrate := int((screen.Width * screen.Height * 5) / 3)

				return fmt.Sprintf(
					"ximagesrc display-name=%s show-pointer=false use-damage=false "+
						"! video/x-raw,framerate=25/1 "+
						"! videoconvert "+
						"! queue "+
						"! vp8enc end-usage=cbr target-bitrate=%d cpu-used=16 threads=4 deadline=100000 undershoot=95 error-resilient=partitions keyframe-max-dist=15 auto-alt-ref=true min-quantizer=6 max-quantizer=12 "+
						"! appsink name=appsink", config.Display, bitrate,
				)
			}),
			"hq": streamNew(codec.VP8(), func() string {
				screen := desktop.GetScreenSize()
				width := int(math.Ceil(float64(screen.Width)/6) * 5)
				height := int(math.Ceil(float64(screen.Height)/6) * 5)
				bitrate := int((width * height * 5) / 3)

				return fmt.Sprintf(
					"ximagesrc display-name=%s show-pointer=false use-damage=false "+
						"! video/x-raw,framerate=25/1 "+
						"! videoconvert "+
						"! queue "+
						"! videoscale "+
						"! video/x-raw,width=%d,height=%d "+
						"! queue "+
						"! vp8enc end-usage=cbr target-bitrate=%d cpu-used=16 threads=4 deadline=100000 undershoot=95 error-resilient=partitions keyframe-max-dist=15 auto-alt-ref=true min-quantizer=6 max-quantizer=12 "+
						"! appsink name=appsink", config.Display, width, height, bitrate,
				)
			}),
			"mq": streamNew(codec.VP8(), func() string {
				screen := desktop.GetScreenSize()
				width := int(math.Ceil(float64(screen.Width)/6) * 4)
				height := int(math.Ceil(float64(screen.Height)/6) * 4)
				bitrate := int((width * height * 5) / 3)

				return fmt.Sprintf(
					"ximagesrc display-name=%s show-pointer=false use-damage=false "+
						"! video/x-raw,framerate=125/10 "+
						"! videoconvert "+
						"! queue "+
						"! videoscale "+
						"! video/x-raw,width=%d,height=%d "+
						"! queue "+
						"! vp8enc end-usage=cbr target-bitrate=%d cpu-used=16 threads=4 deadline=100000 undershoot=95 error-resilient=partitions keyframe-max-dist=15 auto-alt-ref=true min-quantizer=12 max-quantizer=24 "+
						"! appsink name=appsink", config.Display, width, height, bitrate,
				)
			}),
			"lq": streamNew(codec.VP8(), func() string {
				screen := desktop.GetScreenSize()
				width := int(math.Ceil(float64(screen.Width)/6) * 3)
				height := int(math.Ceil(float64(screen.Height)/6) * 3)
				bitrate := int((width * height * 5) / 3)

				return fmt.Sprintf(
					"ximagesrc display-name=%s show-pointer=false use-damage=false "+
						"! video/x-raw,framerate=125/10 "+
						"! videoconvert "+
						"! queue "+
						"! videoscale "+
						"! video/x-raw,width=%d,height=%d "+
						"! queue "+
						"! vp8enc end-usage=cbr target-bitrate=%d cpu-used=16 threads=4 deadline=100000 undershoot=95 error-resilient=partitions keyframe-max-dist=15 auto-alt-ref=true min-quantizer=12 max-quantizer=24 "+
						"! appsink name=appsink", config.Display, width, height, bitrate,
				)
			}),
		},
		videoIDs: []string{"hd", "hq", "mq", "lq"},
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
