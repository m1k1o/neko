package capture

import (
	"errors"
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/capture/gst"
	"m1k1o/neko/internal/config"
	"m1k1o/neko/internal/types"

	go_gst "github.com/go-gst/go-gst/gst"
	go_gst_app "github.com/go-gst/go-gst/gst/app"
)

type CaptureManagerCtx struct {
	logger  zerolog.Logger
	desktop types.DesktopManager
	config  *config.Capture // Store main capture config
	wsConfig *config.WebSocket // Store websocket config

	// sinks
	broadcast *BroacastManagerCtx
	audio     *StreamSinkManagerCtx
	video     *StreamSinkManagerCtx

	// websocket fMP4 pipeline
	wsPipeline   *gst.Pipeline
	wsAppSink    *go_gst_app.Sink
	wsMux        sync.Mutex
	wsClientCnt  int
	wsPipelineMu sync.Mutex
}

func New(desktop types.DesktopManager, config *config.Capture, wsConfig *config.WebSocket) *CaptureManagerCtx {
	logger := log.With().Str("module", "capture").Logger()

	return &CaptureManagerCtx{
		logger:  logger,
		desktop: desktop,
		config:  config,  // Store capture config
		wsConfig: wsConfig, // Store websocket config

		// sinks
		broadcast: broadcastNew(func(url string) (string, error) {
			return NewBroadcastPipeline(config.AudioDevice, config.Display, config.BroadcastPipeline, url)
		}, config.BroadcastUrl, config.BroadcastAutostart),
		audio: streamSinkNew(config.AudioCodec, func() (string, error) {
			return NewAudioPipeline(config.AudioCodec, config.AudioDevice, config.AudioPipeline, config.AudioBitrate)
		}, "audio"),
		video: streamSinkNew(config.VideoCodec, func() (string, error) {
			// use screen fps as default
			fps := desktop.GetScreenSize().Rate
			// if max fps is set, cap it to that value
			if config.VideoMaxFPS > 0 && config.VideoMaxFPS < fps {
				fps = config.VideoMaxFPS
			}
			return NewVideoPipeline(config.VideoCodec, config.Display, config.VideoPipeline, fps, config.VideoBitrate, config.VideoHWEnc)
		}, "video"),

		// ws fields initialized to zero values (nil, nil, zero mutex, 0)
	}
}

func (manager *CaptureManagerCtx) Start() {
	if manager.broadcast.Started() {
		if err := manager.broadcast.createPipeline(); err != nil {
			manager.logger.Panic().Err(err).Msg("unable to create broadcast pipeline")
		}
	}

	go gst.RunMainLoop()
	go func() {
		for {
			before, ok := <-manager.desktop.GetScreenSizeChangeChannel()
			if !ok {
				manager.logger.Info().Msg("screen size change channel was closed")
				return
			}

			if before {
				// before screen size change, we need to destroy all pipelines

				if manager.video.Started() {
					manager.video.destroyPipeline()
				}

				if manager.broadcast.Started() {
					manager.broadcast.destroyPipeline()
				}

				// Destroy WebSocket pipeline before screen change
				manager.DestroyWebSocketPipeline()

			} else {
				// after screen size change, we need to recreate all pipelines

				if manager.video.Started() {
					err := manager.video.createPipeline()
					if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
						manager.logger.Panic().Err(err).Msg("unable to recreate video pipeline")
					}
				}

				if manager.broadcast.Started() {
					err := manager.broadcast.createPipeline()
					if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
						manager.logger.Panic().Err(err).Msg("unable to recreate broadcast pipeline")
					}
				}

				// Recreate WebSocket pipeline if clients were connected
				manager.wsPipelineMu.Lock()
				clientCount := manager.wsClientCnt
				manager.wsPipelineMu.Unlock()

				if clientCount > 0 {
					manager.logger.Info().Msg("recreating WebSocket pipeline after screen change")
					_, err := manager.EnsureWebSocketPipeline() // Recreate pipeline
					if err != nil {
						// TODO: Handle this more gracefully? Maybe retry?
						manager.logger.Error().Err(err).Msg("failed to recreate WebSocket pipeline after screen change")
					}
				}
			}
		}
	}()
}

func (manager *CaptureManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("shutdown")

	manager.broadcast.shutdown()

	manager.audio.shutdown()
	manager.video.shutdown()

	// Destroy WebSocket pipeline on shutdown
	manager.DestroyWebSocketPipeline()

	gst.QuitMainLoop()

	return nil
}

func (manager *CaptureManagerCtx) Broadcast() types.BroadcastManager {
	return manager.broadcast
}

func (manager *CaptureManagerCtx) Audio() types.StreamSinkManager {
	return manager.audio
}

func (manager *CaptureManagerCtx) Video() types.StreamSinkManager {
	return manager.video
}

// EnsureWebSocketPipeline creates the fMP4 pipeline if it doesn't exist,
// increments the client counter, and returns the appsink.
func (manager *CaptureManagerCtx) EnsureWebSocketPipeline() (*go_gst_app.Sink, error) {
	manager.wsPipelineMu.Lock()
	defer manager.wsPipelineMu.Unlock()

	if manager.wsPipeline != nil {
		manager.wsClientCnt++
		manager.logger.Debug().Int("count", manager.wsClientCnt).Msg("incremented WebSocket client count")
		return manager.wsAppSink, nil
	}

	manager.logger.Info().Msg("creating WebSocket pipeline")

	// Get current screen settings
	screenSize := manager.desktop.GetScreenSize()
	fps := screenSize.Rate
	// if max fps is set, cap it to that value
	if manager.config.VideoMaxFPS > 0 && manager.config.VideoMaxFPS < fps {
		fps = manager.config.VideoMaxFPS
	}

	pipelineStr, err := NewWebSocketPipeline(
		manager.wsConfig.VideoCodec,
		manager.wsConfig.AudioCodec,
		manager.config.Display,
		manager.config.AudioDevice,
		"", // No custom pipeline support for WS yet
		fps,
		manager.wsConfig.VideoBitrate,
		manager.wsConfig.AudioBitrate,
		manager.config.VideoHWEnc,
		manager.wsConfig.FragmentDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build WebSocket pipeline string: %w", err)
	}

	manager.wsPipeline, err = gst.CreatePipeline(pipelineStr)
	if err != nil {
		return nil, err
	}

	// Call the wrapper's GetAppSink (no arguments)
	manager.wsAppSink = manager.wsPipeline.GetAppSink()
	if manager.wsAppSink == nil {
		manager.wsPipeline.Stop() // Stop calls Destroy internally now
		manager.wsPipeline = nil // Stop should already delete from map, but good practice
		return nil, fmt.Errorf("failed to get appsink element from WebSocket pipeline wrapper using Ctx field")
	}

	// Set the pipeline to playing using the wrapper's method
	manager.wsPipeline.Play()

	manager.wsClientCnt = 1
	manager.logger.Info().Msg("WebSocket pipeline created and started")
	return manager.wsAppSink, nil
}

// ReleaseWebSocketPipeline decrements the client counter and destroys the pipeline
// if the counter reaches zero.
func (manager *CaptureManagerCtx) ReleaseWebSocketPipeline() {
	manager.wsPipelineMu.Lock()
	defer manager.wsPipelineMu.Unlock()

	if manager.wsPipeline == nil {
		manager.logger.Warn().Msg("ReleaseWebSocketPipeline called but pipeline is already nil")
		return
	}

	manager.wsClientCnt--
	manager.logger.Debug().Int("count", manager.wsClientCnt).Msg("decremented WebSocket client count")

	if manager.wsClientCnt <= 0 {
		manager.logger.Info().Msg("destroying WebSocket pipeline as last client disconnected")
		manager.destroyWebSocketPipelineInternal()
	}
}

// DestroyWebSocketPipeline force destroys the pipeline, regardless of client count.
// Used for shutdown and screen changes.
func (manager *CaptureManagerCtx) DestroyWebSocketPipeline() {
	manager.wsPipelineMu.Lock()
	defer manager.wsPipelineMu.Unlock()

	if manager.wsPipeline == nil {
		return // Already destroyed
	}

	manager.logger.Info().Msg("force destroying WebSocket pipeline")
	manager.destroyWebSocketPipelineInternal()
}

// destroyWebSocketPipelineInternal performs the actual destruction. MUST be called with wsPipelineMu locked.
func (manager *CaptureManagerCtx) destroyWebSocketPipelineInternal() {
	if manager.wsPipeline == nil {
		return
	}
	// Stop now calls the C destroy function and removes from map in gst.go
	manager.wsPipeline.Stop()

	manager.wsPipeline = nil // Ensure Go reference is cleared
	manager.wsAppSink = nil
	manager.wsClientCnt = 0 // Reset count on force destroy
	manager.logger.Debug().Msg("WebSocket pipeline destroyed internal")
}
