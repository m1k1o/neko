package webrtc

import (
	"bytes"
	"encoding/binary"
	"sync"
	"time"

	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog"

	"github.com/demodesk/neko/internal/config"
	"github.com/demodesk/neko/internal/webrtc/payload"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
	"github.com/demodesk/neko/pkg/utils"
)

type WebRTCPeerCtx struct {
	mu         sync.Mutex
	logger     zerolog.Logger
	session    types.Session
	metrics    *metrics
	connection *webrtc.PeerConnection
	// bandwidth estimator
	estimator     cc.BandwidthEstimator
	estimateTrend *utils.TrendDetector
	// stream selectors
	videoSelector types.StreamSelectorManager
	// tracks & channels
	audioTrack  *Track
	videoTrack  *Track
	dataChannel *webrtc.DataChannel
	rtcpChannel chan []rtcp.Packet
	// config
	iceTrickle      bool
	estimatorConfig config.WebRTCEstimator
	videoAuto       bool
}

//
// connection
//

func (peer *WebRTCPeerCtx) CreateOffer(ICERestart bool) (*webrtc.SessionDescription, error) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	offer, err := peer.connection.CreateOffer(&webrtc.OfferOptions{
		ICERestart: ICERestart,
	})
	if err != nil {
		return nil, err
	}

	return peer.setLocalDescription(offer)
}

func (peer *WebRTCPeerCtx) CreateAnswer() (*webrtc.SessionDescription, error) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	answer, err := peer.connection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	return peer.setLocalDescription(answer)
}

func (peer *WebRTCPeerCtx) setLocalDescription(description webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	if !peer.iceTrickle {
		// Create channel that is blocked until ICE Gathering is complete
		gatherComplete := webrtc.GatheringCompletePromise(peer.connection)

		if err := peer.connection.SetLocalDescription(description); err != nil {
			return nil, err
		}

		<-gatherComplete
	} else {
		if err := peer.connection.SetLocalDescription(description); err != nil {
			return nil, err
		}
	}

	return peer.connection.LocalDescription(), nil
}

func (peer *WebRTCPeerCtx) SetRemoteDescription(desc webrtc.SessionDescription) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return peer.connection.SetRemoteDescription(desc)
}

func (peer *WebRTCPeerCtx) SetCandidate(candidate webrtc.ICECandidateInit) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return peer.connection.AddICECandidate(candidate)
}

// TODO: Add shutdown function?
func (peer *WebRTCPeerCtx) Destroy() {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	err := peer.connection.Close()
	peer.logger.Err(err).Msg("peer connection destroyed")
}

func (peer *WebRTCPeerCtx) estimatorReader() {
	conf := peer.estimatorConfig

	// if estimator is not in debug mode, use a nop logger
	var debugLogger zerolog.Logger
	if conf.Debug {
		debugLogger = peer.logger.With().Str("component", "estimator").Logger().Level(zerolog.DebugLevel)
	} else {
		debugLogger = zerolog.Nop()
	}

	// if estimator is disabled, do nothing
	if peer.estimator == nil {
		return
	}

	// use a ticker to get current client target bitrate
	ticker := time.NewTicker(conf.ReadInterval)
	defer ticker.Stop()

	// since when is the estimate stable/unstable
	stableSince := time.Now() // we asume stable at start
	unstableSince := time.Time{}
	// since when are we neutral but cannot accomodate current bitrate
	// we migt be stalled or estimator just reached zer (very bad connection)
	stalledSince := time.Time{}
	// when was the last upgrade/downgrade
	lastUpgradeTime := time.Time{}
	lastDowngradeTime := time.Time{}

	for range ticker.C {
		targetBitrate := peer.estimator.GetTargetBitrate()
		peer.metrics.SetReceiverEstimatedTargetBitrate(float64(targetBitrate))

		// if peer connection is closed, stop reading
		if peer.connection.ConnectionState() == webrtc.PeerConnectionStateClosed {
			break
		}

		// if estimation is disabled, do nothing
		if !peer.videoAuto || conf.Passive {
			continue
		}

		// get trend direction to decide if we should upgrade or downgrade
		peer.estimateTrend.AddValue(int64(targetBitrate))
		direction := peer.estimateTrend.GetDirection()

		// get current stream bitrate
		stream, ok := peer.videoTrack.Stream()
		if !ok {
			debugLogger.Warn().Msg("looks like we don't have a stream yet, skipping bitrate estimation")
			continue
		}

		// if stream bitrate is 0, we need to wait for some time until we get a valid value
		streamId, streamBitrate := stream.ID(), stream.Bitrate()
		if streamBitrate == 0 {
			debugLogger.Warn().Msg("looks like stream bitrate is 0, we need to wait for some time")
			continue
		}

		// check whats the difference between target and stream bitrate
		diff := float64(targetBitrate) / float64(streamBitrate)

		debugLogger.Info().
			Float64("diff", diff).
			Int("target_bitrate", targetBitrate).
			Uint64("stream_bitrate", streamBitrate).
			Str("direction", direction.String()).
			Msg("got bitrate from estimator")

		// if we can accomodate current stream or we are not netural anymore,
		// we are not stalled so we reset the stalled time
		if direction != utils.TrendDirectionNeutral || diff > 1+conf.DiffThreshold {
			stalledSince = time.Now()
		}

		// if we are neutral and stalled for too long, we might be congesting
		stalled := direction == utils.TrendDirectionNeutral && time.Since(stalledSince) > conf.StalledDuration
		if stalled {
			debugLogger.Warn().
				Time("stalled_since", stalledSince).
				Msgf("it looks like we are stalled")
		}

		// if we have an downward trend or are stalled, we might be congesting
		if direction == utils.TrendDirectionDownward || stalled {
			// we reset the stable time because we are congesting
			stableSince = time.Now()

			// if we downgraded recently, we wait for some more time
			if time.Since(lastDowngradeTime) < conf.DowngradeBackoff {
				debugLogger.Debug().
					Time("last_downgrade", lastDowngradeTime).
					Msgf("downgraded recently, waiting for at least %v", conf.DowngradeBackoff)
				continue
			}

			// if we are not unstable but we fluctuate we should wait for some more time
			if time.Since(unstableSince) < conf.UnstableDuration {
				debugLogger.Debug().
					Time("unstable_since", unstableSince).
					Msgf("we are not unstable long enough, waiting for at least %v", conf.UnstableDuration)
				continue
			}

			// if we still have a big difference between target and stream bitrate, we wait for some more time
			if conf.DiffThreshold >= 0 && diff > 1+conf.DiffThreshold {
				debugLogger.Debug().
					Float64("diff", diff).
					Float64("threshold", conf.DiffThreshold).
					Msgf("we still have a big difference between target and stream bitrate, " +
						"therefore we still should be able to accomodate current stream")
				continue
			}

			err := peer.SetVideo(types.StreamSelector{
				ID:   streamId,
				Type: types.StreamSelectorTypeLower,
			})
			if err != nil && err != types.ErrWebRTCStreamNotFound {
				peer.logger.Warn().Err(err).Msg("failed to downgrade video stream")
			}
			lastDowngradeTime = time.Now()

			if err == types.ErrWebRTCStreamNotFound {
				debugLogger.Info().Msg("looks like we are already on the lowest stream")
			} else {
				debugLogger.Info().Msg("downgraded video stream")
			}
			continue
		}

		// we reset the unstable time because we are not congesting
		unstableSince = time.Now()

		// if we have a neutral or upward trend, that means our estimate is stable
		// if we are on the highest stream, we don't need to do anything
		// but if there is a higher stream, we should try to upgrade and see if it works

		// if we upgraded recently, we wait for some more time
		if time.Since(lastUpgradeTime) < conf.UpgradeBackoff {
			debugLogger.Debug().
				Time("last_upgrade", lastUpgradeTime).
				Msgf("upgraded recently, waiting for at least %v", conf.UpgradeBackoff)
			continue
		}

		// if we are not stable for long enough, we wait for some more time
		// because bandwidth estimation might fluctuate
		if time.Since(stableSince) < conf.StableDuration {
			debugLogger.Debug().
				Time("stable_since", stableSince).
				Msgf("we are not stable long enough, waiting for at least %v", conf.StableDuration)
			continue
		}

		// upgrade only if estimated bitrate passed the threshold
		if conf.DiffThreshold >= 0 && diff < 1+conf.DiffThreshold {
			debugLogger.Debug().
				Float64("diff", diff).
				Float64("threshold", conf.DiffThreshold).
				Msgf("looks like we don't have enough bitrate to accomodate higher stream, " +
					"therefore we should wait for some more time")
			continue
		}

		err := peer.SetVideo(types.StreamSelector{
			ID:   streamId,
			Type: types.StreamSelectorTypeHigher,
		})
		if err != nil && err != types.ErrWebRTCStreamNotFound {
			peer.logger.Warn().Err(err).Msg("failed to upgrade video stream")
		}
		lastUpgradeTime = time.Now()

		if err == types.ErrWebRTCStreamNotFound {
			debugLogger.Info().Msg("looks like we are already on the highest stream")
		} else {
			debugLogger.Info().Msg("upgraded video stream")
		}
	}
}

//
// video
//

func (peer *WebRTCPeerCtx) SetVideo(selector types.StreamSelector) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	// get requested video stream from selector
	stream, ok := peer.videoSelector.GetStream(selector)
	if !ok {
		return types.ErrWebRTCStreamNotFound
	}

	// set video stream to track
	changed, err := peer.videoTrack.SetStream(stream)
	if err != nil {
		return err
	}

	// if video stream was already set, do nothing
	if !changed {
		return nil
	}

	videoID := stream.ID()
	peer.metrics.SetVideoID(videoID)

	peer.logger.Info().Str("video_id", videoID).Msg("set video")

	go peer.session.Send(
		event.SIGNAL_VIDEO,
		message.SignalVideo{
			Video: videoID,
			Auto:  peer.videoAuto,
		})

	return nil
}

func (peer *WebRTCPeerCtx) VideoID() (string, bool) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	stream, ok := peer.videoTrack.Stream()
	if !ok {
		return "", false
	}

	return stream.ID(), true
}

func (peer *WebRTCPeerCtx) SetPaused(isPaused bool) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	peer.logger.Info().Bool("is_paused", isPaused).Msg("set paused")
	peer.videoTrack.SetPaused(isPaused)
	peer.audioTrack.SetPaused(isPaused)
	return nil
}

func (peer *WebRTCPeerCtx) Paused() bool {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return peer.videoTrack.Paused() || peer.audioTrack.Paused()
}

func (peer *WebRTCPeerCtx) SetVideoAuto(videoAuto bool) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	// if estimator is enabled and is not passive, enable video auto bitrate
	if peer.estimator != nil && !peer.estimatorConfig.Passive {
		peer.logger.Info().Bool("video_auto", videoAuto).Msg("set video auto")
		peer.videoAuto = videoAuto
	} else {
		peer.logger.Warn().Msg("estimator is disabled or in passive mode, cannot change video auto")
		peer.videoAuto = false // ensure video auto is disabled
	}
}

func (peer *WebRTCPeerCtx) VideoAuto() bool {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return peer.videoAuto
}

//
// data channel
//

func (peer *WebRTCPeerCtx) SendCursorPosition(x, y int) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	// do not send cursor position to host
	if peer.session.IsHost() {
		return nil
	}

	header := payload.Header{
		Event:  payload.OP_CURSOR_POSITION,
		Length: 7,
	}

	data := payload.CursorPosition{
		X: uint16(x),
		Y: uint16(y),
	}

	buffer := &bytes.Buffer{}

	if err := binary.Write(buffer, binary.BigEndian, header); err != nil {
		return err
	}

	if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
		return err
	}

	return peer.dataChannel.Send(buffer.Bytes())
}

func (peer *WebRTCPeerCtx) SendCursorImage(cur *types.CursorImage, img []byte) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	header := payload.Header{
		Event:  payload.OP_CURSOR_IMAGE,
		Length: uint16(11 + len(img)),
	}

	data := payload.CursorImage{
		Width:  cur.Width,
		Height: cur.Height,
		Xhot:   cur.Xhot,
		Yhot:   cur.Yhot,
	}

	buffer := &bytes.Buffer{}

	if err := binary.Write(buffer, binary.BigEndian, header); err != nil {
		return err
	}

	if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
		return err
	}

	if err := binary.Write(buffer, binary.BigEndian, img); err != nil {
		return err
	}

	return peer.dataChannel.Send(buffer.Bytes())
}
