package capture

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

type CaptureManagerCtx struct {
	logger  zerolog.Logger
	desktop types.DesktopManager
	config  *config.Capture

	// sinks
	broadcast  *BroacastManagerCtx
	screencast *ScreencastManagerCtx
	audio      *StreamSinkManagerCtx
	video      *StreamSelectorManagerCtx

	// sources
	webcam     *StreamSrcManagerCtx
	microphone *StreamSrcManagerCtx
}

func replaceCapturePlaceholders(pipeline string, captureConfig *config.Capture) (string, error) {
	pipeline = strings.ReplaceAll(pipeline, "{display}", captureConfig.Display)
	if captureConfig.WindowWidth > 0 && captureConfig.WindowHeight > 0 {
		for _, placeholder := range []string{"{window_x}", "{window_y}"} {
			if !strings.Contains(pipeline, placeholder) {
				return "", fmt.Errorf("custom capture pipeline must contain %s while targeting an X11 window region", placeholder)
			}
		}
		hasDimensions := strings.Contains(pipeline, "{window_width}") && strings.Contains(pipeline, "{window_height}")
		hasEnds := strings.Contains(pipeline, "{window_end_x}") && strings.Contains(pipeline, "{window_end_y}")
		if !hasDimensions && !hasEnds {
			return "", fmt.Errorf("custom capture pipeline must contain window width/height or end-coordinate placeholders while targeting an X11 window region")
		}
	}
	if captureConfig.WindowID != 0 && captureConfig.WindowWidth == 0 && !strings.Contains(pipeline, "{window_id}") {
		return "", fmt.Errorf("custom capture pipeline must contain {window_id} while targeting an X11 window")
	}
	replacements := map[string]string{
		"{window_id}":     fmt.Sprintf("%d", captureConfig.WindowID),
		"{window_x}":      fmt.Sprintf("%d", captureConfig.WindowX),
		"{window_y}":      fmt.Sprintf("%d", captureConfig.WindowY),
		"{window_width}":  fmt.Sprintf("%d", captureConfig.WindowWidth),
		"{window_height}": fmt.Sprintf("%d", captureConfig.WindowHeight),
		"{window_end_x}":  fmt.Sprintf("%d", captureConfig.WindowX+captureConfig.WindowWidth-1),
		"{window_end_y}":  fmt.Sprintf("%d", captureConfig.WindowY+captureConfig.WindowHeight-1),
	}
	for placeholder, value := range replacements {
		pipeline = strings.ReplaceAll(pipeline, placeholder, value)
	}
	return pipeline, nil
}

func xImageSource(captureConfig *config.Capture, showPointer bool) string {
	source := fmt.Sprintf("ximagesrc display-name=%s", captureConfig.Display)
	if captureConfig.WindowWidth > 0 && captureConfig.WindowHeight > 0 {
		source += fmt.Sprintf(
			" startx=%d starty=%d endx=%d endy=%d",
			captureConfig.WindowX,
			captureConfig.WindowY,
			captureConfig.WindowX+captureConfig.WindowWidth-1,
			captureConfig.WindowY+captureConfig.WindowHeight-1,
		)
	} else if captureConfig.WindowID != 0 {
		source += fmt.Sprintf(" xid=%d", captureConfig.WindowID)
	}
	return fmt.Sprintf("%s show-pointer=%v use-damage=false", source, showPointer)
}

func New(desktop types.DesktopManager, config *config.Capture) *CaptureManagerCtx {
	logger := log.With().Str("module", "capture").Logger()
	if (config.WindowWidth > 0) != (config.WindowHeight > 0) ||
		config.WindowX < 0 || config.WindowY < 0 ||
		(config.WindowID != 0 && config.WindowWidth > 0) {
		logger.Panic().Msg("invalid X11 window capture configuration")
	}

	videos := map[string]types.StreamSinkManager{}
	for video_id, cnf := range config.VideoPipelines {
		pipelineConf := cnf

		createPipeline := func() (string, error) {
			if pipelineConf.GstPipeline != "" {
				return replaceCapturePlaceholders(pipelineConf.GstPipeline, config)
			}

			screen := desktop.GetScreenSize()
			pipeline, err := pipelineConf.GetPipeline(screen)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf(
				xImageSource(config, pipelineConf.ShowPointer)+" "+
					"%s ! appsink name=appsink", pipeline,
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
		config:  config,

		// sinks
		broadcast: broadcastNew(func(url string) (string, error) {
			if config.BroadcastPipeline != "" {
				var pipeline = config.BroadcastPipeline
				if hostname, err := os.Hostname(); err == nil {
					// replace {hostname} with valid hostname
					pipeline = strings.Replace(pipeline, "{hostname}", hostname, 1)
				}
				// replace capture source placeholders
				var err error
				pipeline, err = replaceCapturePlaceholders(pipeline, config)
				if err != nil {
					return "", err
				}
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
					xImageSource(config, true)+" "+
					"! video/x-raw "+
					"! videoconvert "+
					"! queue "+
					"! x264enc threads=4 bitrate=%d key-int-max=15 byte-stream=true tune=zerolatency speed-preset=%s "+
					"! mux.", url, config.AudioDevice, config.BroadcastAudioBitrate*1000, config.BroadcastVideoBitrate, config.BroadcastPreset,
			), nil
		}, config.BroadcastUrl, config.BroadcastAutostart),
		screencast: screencastNew(config.ScreencastEnabled, func() string {
			if config.ScreencastPipeline != "" {
				pipeline, err := replaceCapturePlaceholders(config.ScreencastPipeline, config)
				if err != nil {
					logger.Panic().Err(err).Msg("invalid screencast pipeline")
				}
				return pipeline
			}

			return fmt.Sprintf(
				xImageSource(config, true)+" "+
					"! video/x-raw,framerate=%s "+
					"! videoconvert "+
					"! queue "+
					"! jpegenc quality=%s "+
					"! appsink name=appsink", config.ScreencastRate, config.ScreencastQuality,
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
					"! queue max-size-buffers=5 leaky=downstream "+
					"! %s "+
					"! appsink name=appsink", config.AudioDevice, config.AudioCodec.Pipeline,
			), nil
		}, "audio"),
		video: streamSelectorNew(config.VideoCodec, videos, config.VideoIDs),

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
		manager.video.destroyPipelines()

		if manager.broadcast.Started() {
			manager.broadcast.destroyPipeline()
		}

		if manager.screencast.Started() {
			manager.screencast.destroyPipeline()
		}
	})

	manager.desktop.OnAfterScreenSizeChange(func() {
		err := manager.video.recreatePipelines()
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

func (manager *CaptureManagerCtx) Video() types.StreamSelectorManager {
	return manager.video
}

func (manager *CaptureManagerCtx) Webcam() types.StreamSrcManager {
	return manager.webcam
}

func (manager *CaptureManagerCtx) Microphone() types.StreamSrcManager {
	return manager.microphone
}
