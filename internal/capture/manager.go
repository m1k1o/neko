package capture

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/internal/config"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/codec"
)

type CaptureManagerCtx struct {
	logger  zerolog.Logger
	desktop types.DesktopManager

	// sinks
	broadcast  *BroacastManagerCtx
	screencast *ScreencastManagerCtx
	audio      *StreamSinkManagerCtx
	video      *BucketsManagerCtx

	// sources
	webcam     *StreamSrcManagerCtx
	microphone *StreamSrcManagerCtx
}

func New(desktop types.DesktopManager, config *config.Capture) *CaptureManagerCtx {
	logger := log.With().Str("module", "capture").Logger()

	videos := map[string]*StreamSinkManagerCtx{}
	for video_id, cnf := range config.VideoPipelines {
		pipelineConf := cnf

		createPipeline := func() (string, error) {
			if pipelineConf.GstPipeline != "" {
				// replace {display} with valid display
				return strings.Replace(pipelineConf.GstPipeline, "{display}", config.Display, 1), nil
			}

			screen := desktop.GetScreenSize()
			pipeline, err := pipelineConf.GetPipeline(*screen)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf(
				"ximagesrc display-name=%s show-pointer=false use-damage=false "+
					"%s ! appsink name=appsink", config.Display, pipeline,
			), nil
		}

		// trigger function to catch evaluation errors at startup
		pipeline, err := createPipeline()
		if err != nil {
			logger.Panic().Err(err).
				Str("video_id", video_id).
				Msg("failed to create video pipeline")
		}

		logger.Info().
			Str("video_id", video_id).
			Str("pipeline", pipeline).
			Msg("syntax check for video stream pipeline passed")

		// append to videos
		videos[video_id] = streamSinkNew(config.VideoCodec, createPipeline, video_id)
	}

	return &CaptureManagerCtx{
		logger:  logger,
		desktop: desktop,

		// sinks
		broadcast: broadcastNew(func(url string) (string, error) {
			if config.BroadcastPipeline != "" {
				var pipeline = config.BroadcastPipeline
				// replace {display} with valid display
				pipeline = strings.Replace(pipeline, "{display}", config.Display, 1)
				// replace {device} with valid device
				pipeline = strings.Replace(pipeline, "{device}", config.AudioDevice, 1)
				// replace {url} with valid URL
				return strings.Replace(pipeline, "{url}", url, 1), nil
			}

			return fmt.Sprintf(
				"flvmux name=mux ! rtmpsink location='%s live=1' "+
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
					"! mux.", url, config.AudioDevice, config.BroadcastAudioBitrate*1000, config.Display, config.BroadcastVideoBitrate, config.BroadcastPreset,
			), nil
		}, config.BroadcastUrl),
		screencast: screencastNew(config.ScreencastEnabled, func() string {
			if config.ScreencastPipeline != "" {
				// replace {display} with valid display
				return strings.Replace(config.ScreencastPipeline, "{display}", config.Display, 1)
			}

			return fmt.Sprintf(
				"ximagesrc display-name=%s show-pointer=true use-damage=false "+
					"! video/x-raw,framerate=%s "+
					"! videoconvert "+
					"! queue "+
					"! jpegenc quality=%s "+
					"! appsink name=appsink", config.Display, config.ScreencastRate, config.ScreencastQuality,
			)
		}()),

		audio: streamSinkNew(config.AudioCodec, func() (string, error) {
			if config.AudioPipeline != "" {
				// replace {device} with valid device
				return strings.Replace(config.AudioPipeline, "{device}", config.AudioDevice, 1), nil
			}

			return fmt.Sprintf(
				"pulsesrc device=%s "+
					"! audio/x-raw,channels=2 "+
					"! audioconvert "+
					"! queue "+
					"! %s "+
					"! appsink name=appsink", config.AudioDevice, config.AudioCodec.Pipeline,
			), nil
		}, "audio"),
		video: bucketsNew(config.VideoCodec, videos, config.VideoIDs),

		// sources
		webcam: streamSrcNew(config.WebcamEnabled, map[string]string{
			codec.VP8().Name: "appsrc format=time is-live=true do-timestamp=true name=appsrc " +
				fmt.Sprintf("! application/x-rtp, payload=%d, encoding-name=VP8-DRAFT-IETF-01 ", codec.VP8().PayloadType) +
				"! rtpvp8depay " +
				"! decodebin " +
				"! videoconvert " +
				"! videorate " +
				"! videoscale " +
				fmt.Sprintf("! video/x-raw,width=%d,height=%d ", config.WebcamWidth, config.WebcamHeight) +
				"! identity drop-allocation=true " +
				fmt.Sprintf("! v4l2sink sync=false device=%s", config.WebcamDevice),
			// TODO: Test this pipeline.
			codec.VP9().Name: "appsrc format=time is-live=true do-timestamp=true name=appsrc " +
				"! application/x-rtp " +
				"! rtpvp9depay " +
				"! decodebin " +
				"! videoconvert " +
				"! videorate " +
				"! videoscale " +
				fmt.Sprintf("! video/x-raw,width=%d,height=%d ", config.WebcamWidth, config.WebcamHeight) +
				"! identity drop-allocation=true " +
				fmt.Sprintf("! v4l2sink sync=false device=%s", config.WebcamDevice),
			// TODO: Test this pipeline.
			codec.H264().Name: "appsrc format=time is-live=true do-timestamp=true name=appsrc " +
				"! application/x-rtp " +
				"! rtph264depay " +
				"! decodebin " +
				"! videoconvert " +
				"! videorate " +
				"! videoscale " +
				fmt.Sprintf("! video/x-raw,width=%d,height=%d ", config.WebcamWidth, config.WebcamHeight) +
				"! identity drop-allocation=true " +
				fmt.Sprintf("! v4l2sink sync=false device=%s", config.WebcamDevice),
		}, "webcam"),
		microphone: streamSrcNew(config.MicrophoneEnabled, map[string]string{
			codec.Opus().Name: "appsrc format=time is-live=true do-timestamp=true name=appsrc " +
				fmt.Sprintf("! application/x-rtp, payload=%d, encoding-name=OPUS ", codec.Opus().PayloadType) +
				"! rtpopusdepay " +
				"! decodebin " +
				fmt.Sprintf("! pulsesink device=%s", config.MicrophoneDevice),
			// TODO: Test this pipeline.
			codec.G722().Name: "appsrc format=time is-live=true do-timestamp=true name=appsrc " +
				"! application/x-rtp clock-rate=8000 " +
				"! rtpg722depay " +
				"! decodebin " +
				fmt.Sprintf("! pulsesink device=%s", config.MicrophoneDevice),
		}, "microphone"),
	}
}

func (manager *CaptureManagerCtx) Start() {
	if manager.broadcast.Started() {
		if err := manager.broadcast.createPipeline(); err != nil {
			manager.logger.Panic().Err(err).Msg("unable to create broadcast pipeline")
		}
	}

	manager.desktop.OnBeforeScreenSizeChange(func() {
		manager.video.destroyAll()

		if manager.broadcast.Started() {
			manager.broadcast.destroyPipeline()
		}

		if manager.screencast.Started() {
			manager.screencast.destroyPipeline()
		}
	})

	manager.desktop.OnAfterScreenSizeChange(func() {
		err := manager.video.recreateAll()
		if err != nil {
			manager.logger.Panic().Err(err).Msg("unable to recreate video pipelines")
		}

		if manager.broadcast.Started() {
			err := manager.broadcast.createPipeline()
			if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
				manager.logger.Panic().Err(err).Msg("unable to recreate broadcast pipeline")
			}
		}

		if manager.screencast.Started() {
			err := manager.screencast.createPipeline()
			if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
				manager.logger.Panic().Err(err).Msg("unable to recreate screencast pipeline")
			}
		}
	})
}

func (manager *CaptureManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("shutdown")

	manager.broadcast.shutdown()
	manager.screencast.shutdown()

	manager.audio.shutdown()
	manager.video.shutdown()

	manager.webcam.shutdown()
	manager.microphone.shutdown()

	return nil
}

func (manager *CaptureManagerCtx) Broadcast() types.BroadcastManager {
	return manager.broadcast
}

func (manager *CaptureManagerCtx) Screencast() types.ScreencastManager {
	return manager.screencast
}

func (manager *CaptureManagerCtx) Audio() types.StreamSinkManager {
	return manager.audio
}

func (manager *CaptureManagerCtx) Video() types.BucketsManager {
	return manager.video
}

func (manager *CaptureManagerCtx) Webcam() types.StreamSrcManager {
	return manager.webcam
}

func (manager *CaptureManagerCtx) Microphone() types.StreamSrcManager {
	return manager.microphone
}
