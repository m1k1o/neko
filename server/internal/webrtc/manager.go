package webrtc

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"n.eko.moe/neko/internal/gst"
)

func NewManager(password string) (*WebRTCManager, error) {
	engine := webrtc.MediaEngine{}

	videoCodec := webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000)
	video, err := webrtc.NewTrack(webrtc.DefaultPayloadTypeVP8, rand.Uint32(), "stream", "stream", videoCodec)
	if err != nil {
		return nil, err
	}
	engine.RegisterCodec(videoCodec)

	audioCodec := webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000)
	audio, err := webrtc.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "stream", "stream", audioCodec)
	if err != nil {
		return nil, err
	}
	engine.RegisterCodec(audioCodec)

	videoPipeline := gst.CreatePipeline(webrtc.VP8, []*webrtc.Track{video}, "ximagesrc show-pointer=true use-damage=false ! video/x-raw,framerate=30/1 ! videoconvert")
	// ximagesrc xid=0 show-pointer=true ! videoconvert ! queue | videotestsrc

	audioPipeline := gst.CreatePipeline(webrtc.Opus, []*webrtc.Track{audio}, "pulsesrc device=auto_null.monitor ! audioconvert")
	// pulsesrc device=auto_null.monitor ! audioconvert | audiotestsrc
	// gst-launch-1.0 -v pulsesrc device=auto_null.monitor ! audioconvert ! vorbisenc ! oggmux ! filesink location=alsasrc.ogg

	return &WebRTCManager{
		logger:        log.With().Str("service", "webrtc").Logger(),
		engine:        engine,
		api:           webrtc.NewAPI(webrtc.WithMediaEngine(engine)),
		video:         video,
		videoPipeline: videoPipeline,
		audio:         audio,
		audioPipeline: audioPipeline,
		controller:    "",
		password:      password,
		sessions:      make(map[string]*session),
		debounce:      make(map[int]time.Time),
		cleanup:       time.NewTicker(500 * time.Second),
		shutdown:      make(chan bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		config: webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
			SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
		},
	}, nil
}

type WebRTCManager struct {
	logger        zerolog.Logger
	upgrader      websocket.Upgrader
	engine        webrtc.MediaEngine
	api           *webrtc.API
	config        webrtc.Configuration
	password      string
	controller    string
	sessions      map[string]*session
	debounce      map[int]time.Time
	shutdown      chan bool
	cleanup       *time.Ticker
	video         *webrtc.Track
	audio         *webrtc.Track
	videoPipeline *gst.Pipeline
	audioPipeline *gst.Pipeline
}

func (manager *WebRTCManager) Start() error {
	manager.videoPipeline.Start()
	manager.audioPipeline.Start()

	go func() {
		for {
			select {
			case <-manager.shutdown:
				return
			case <-manager.cleanup.C:
				manager.checkKeys()
			}
		}
	}()

	return nil
}

func (manager *WebRTCManager) Shutdown() error {
	manager.cleanup.Stop()
	manager.shutdown <- true
	manager.videoPipeline.Stop()
	manager.audioPipeline.Stop()
	return nil
}
