package capture

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	cfg "github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

type CaptureManagerCtx struct {
	logger  zerolog.Logger
	desktop types.DesktopManager
	config  *cfg.Capture

	// sinks
	broadcast     *BroacastManagerCtx
	screencast    *ScreencastManagerCtx
	audio         *StreamSinkManagerCtx
	video         *StreamSelectorManagerCtx
	videoVariants map[string]*StreamSelectorManagerCtx

	// sources
	webcam     *StreamSrcManagerCtx
	microphone *StreamSrcManagerCtx
}

func New(desktop types.DesktopManager, config *cfg.Capture) *CaptureManagerCtx {
	logger := log.With().Str("module", "capture").Logger()

	createVideoSelector := func(variant cfg.VideoVariant) *StreamSelectorManagerCtx {
		videos := map[string]types.StreamSinkManager{}
		for video_id, cnf := range variant.Pipelines {
			pipelineConf := cnf
			pipelineFallbacks := append([]types.VideoConfig(nil), variant.PipelineFallbacks[video_id]...)

			createPipelineFor := func(videoConfig types.VideoConfig) (string, error) {
				if videoConfig.GstPipeline != "" {
					// replace {display} with valid display
					return strings.Replace(videoConfig.GstPipeline, "{display}", config.Display, 1), nil
				}

				screen := desktop.GetScreenSize()
				pipeline, err := videoConfig.GetPipeline(screen)
				if err != nil {
					return "", err
				}

				return fmt.Sprintf(
					"ximagesrc display-name=%s show-pointer=%v use-damage=false %s ! appsink name=appsink",
					config.Display, videoConfig.ShowPointer, pipeline,
				), nil
			}
			createPipeline := func() (string, error) {
				return createPipelineFor(pipelineConf)
			}
			createPipelineCandidates := func() ([]string, error) {
				candidates := make([]string, 0, 1+len(pipelineFallbacks))
				for _, candidate := range append([]types.VideoConfig{pipelineConf}, pipelineFallbacks...) {
					pipeline, err := createPipelineFor(candidate)
					if err != nil {
						return nil, err
					}
					candidates = append(candidates, pipeline)
				}
				return candidates, nil
			}

			// trigger function to catch evaluation errors at startup
			pipeline, err := createPipeline()
			if err != nil {
				logger.Panic().Err(err).
					Str("video_id", video_id).
					Str("codec", variant.Codec.Name).
					Msg("failed to create video pipeline")
			}

			logger.Info().
				Str("video_id", video_id).
				Str("codec", variant.Codec.Name).
				Str("pipeline", pipeline).
				Msg("syntax check for video stream pipeline passed")

			// append to videos
			if len(pipelineFallbacks) > 0 {
				videos[video_id] = streamSinkNewWithFallback(variant.Codec, createPipeline, createPipelineCandidates, video_id)
			} else {
				videos[video_id] = streamSinkNew(variant.Codec, createPipeline, video_id)
			}
		}
		return streamSelectorNew(variant.Codec, videos, variant.IDs)
	}

	variants := config.VideoVariants
	if len(variants) == 0 {
		variants = map[string]cfg.VideoVariant{
			config.VideoCodec.Name: {
				Codec:             config.VideoCodec,
				IDs:               config.VideoIDs,
				Pipelines:         config.VideoPipelines,
				PipelineFallbacks: config.VideoPipelineFallbacks,
			},
		}
	}

	videoVariants := make(map[string]*StreamSelectorManagerCtx, len(variants))
	for name, variant := range variants {
		videoVariants[name] = createVideoSelector(variant)
	}
	video := videoVariants[config.VideoCodec.Name]
	if video == nil {
		// Profile resolution should always include the primary variant. Keep a
		// deterministic fallback for hand-built test/config instances.
		video = createVideoSelector(config.VideoVariants[config.VideoCodec.Name])
		videoVariants[config.VideoCodec.Name] = video
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
					"ximagesrc display-name=%s show-pointer=%v use-damage=false "+
					"! video/x-raw "+
					"! videoconvert "+
					"! queue "+
					"! x264enc threads=4 bitrate=%d key-int-max=15 byte-stream=true tune=zerolatency speed-preset=%s "+
					"! mux.", url, config.AudioDevice, config.BroadcastAudioBitrate*1000, config.Display, config.VideoShowPointer, config.BroadcastVideoBitrate, config.BroadcastPreset,
			), nil
		}, config.BroadcastUrl, config.BroadcastAutostart),
		screencast: screencastNew(config.ScreencastEnabled, func() string {
			if config.ScreencastPipeline != "" {
				// replace {display} with valid display
				return strings.Replace(config.ScreencastPipeline, "{display}", config.Display, 1)
			}

			return fmt.Sprintf(
				"ximagesrc display-name=%s show-pointer=%v use-damage=false "+
					"! video/x-raw,framerate=%s "+
					"! videoconvert "+
					"! queue "+
					"! jpegenc quality=%s "+
					"! appsink name=appsink", config.Display, config.VideoShowPointer, config.ScreencastRate, config.ScreencastQuality,
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
		video:         video,
		videoVariants: videoVariants,

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
		manager.forEachVideo(func(video *StreamSelectorManagerCtx) {
			video.destroyPipelines()
		})

		if manager.broadcast.Started() {
			manager.broadcast.destroyPipeline()
		}

		if manager.screencast.Started() {
			manager.screencast.destroyPipeline()
		}
	})

	manager.desktop.OnAfterScreenSizeChange(func() {
		var videoErr error
		manager.forEachVideo(func(video *StreamSelectorManagerCtx) {
			if videoErr == nil {
				videoErr = video.recreatePipelines()
			}
		})
		if videoErr != nil {
			manager.logger.Panic().Err(videoErr).Msg("unable to recreate video pipelines")
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
	manager.forEachVideo(func(video *StreamSelectorManagerCtx) {
		video.shutdown()
	})

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

func (manager *CaptureManagerCtx) VideoForCodec(videoCodec codec.RTPCodec) (types.StreamSelectorManager, bool) {
	video, ok := manager.videoVariants[videoCodec.Name]
	if !ok {
		return nil, false
	}
	return video, true
}

func (manager *CaptureManagerCtx) SelectVideoCodec(supported []string) (codec.RTPCodec, bool) {
	if len(supported) == 0 {
		return manager.video.Codec(), true
	}

	available := make(map[string]struct{}, len(supported))
	for _, name := range supported {
		parsed, ok := codec.ParseStr(strings.TrimSpace(name))
		if ok && parsed.IsVideo() {
			available[parsed.Name] = struct{}{}
		}
	}

	// Keep the configured codec as the first choice, then use the codecs in
	// descending Chromium interoperability order.
	order := []string{manager.video.Codec().Name, codec.H264().Name, codec.VP8().Name, codec.AV1().Name, codec.H265().Name}
	seen := make(map[string]struct{}, len(order))
	for _, name := range order {
		if _, duplicate := seen[name]; duplicate {
			continue
		}
		seen[name] = struct{}{}
		if _, ok := available[name]; !ok {
			continue
		}
		if _, ok := manager.videoVariants[name]; ok {
			selected, _ := codec.ParseStr(name)
			return selected, true
		}
	}

	return codec.RTPCodec{}, false
}

func (manager *CaptureManagerCtx) forEachVideo(fn func(*StreamSelectorManagerCtx)) {
	for _, video := range manager.videoVariants {
		fn(video)
	}
}

func (manager *CaptureManagerCtx) Webcam() types.StreamSrcManager {
	return manager.webcam
}

func (manager *CaptureManagerCtx) Microphone() types.StreamSrcManager {
	return manager.microphone
}
