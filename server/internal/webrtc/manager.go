package webrtc

import (
	"math/rand"
	"net/http"

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
	gst.CreatePipeline(webrtc.VP8, []*webrtc.Track{video}, "ximagesrc show-pointer=true use-damage=false ! video/x-raw,framerate=30/1 ! videoconvert").Start()
	engine.RegisterCodec(videoCodec)
	// ximagesrc xid=0 show-pointer=true ! videoconvert ! queue | videotestsrc

	audioCodec := webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000)
	audio, err := webrtc.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "stream", "stream", audioCodec)
	if err != nil {
		return nil, err
	}
	gst.CreatePipeline(webrtc.Opus, []*webrtc.Track{audio}, "pulsesrc device=auto_null.monitor ! audioconvert").Start()
	engine.RegisterCodec(audioCodec)
	// pulsesrc device=auto_null.monitor ! audioconvert | audiotestsrc
	// gst-launch-1.0 -v pulsesrc device=auto_null.monitor ! audioconvert ! vorbisenc ! oggmux ! filesink location=alsasrc.ogg

	return &WebRTCManager{
		logger:     log.With().Str("service", "webrtc").Logger(),
		engine:     engine,
		api:        webrtc.NewAPI(webrtc.WithMediaEngine(engine)),
		video:      video,
		audio:      audio,
		controller: "",
		password:   password,
		sessions:   make(map[string]*session),
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
	logger     zerolog.Logger
	upgrader   websocket.Upgrader
	engine     webrtc.MediaEngine
	api        *webrtc.API
	config     webrtc.Configuration
	password   string
	controller string
	sessions   map[string]*session
	video      *webrtc.Track
	audio      *webrtc.Track
}
