package webrtc

import (
	"fmt"
	"math/rand"

	"github.com/pion/webrtc/v2"
	"n.eko.moe/neko/internal/gst"
)

func (m *WebRTCManager) createTrack(codecName string, pipelineDevice string, pipelineSrc string) (*gst.Pipeline, *webrtc.Track, error) {
	pipeline, err := gst.CreatePipeline(
		codecName,
		pipelineDevice,
		pipelineSrc,
	)

	if err != nil {
		return nil, nil, err
	}

	var codec *webrtc.RTPCodec
	switch codecName {
	case webrtc.VP8:
		codec = webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000)
	case webrtc.VP9:
		codec = webrtc.NewRTPVP9Codec(webrtc.DefaultPayloadTypeVP9, 90000)
	case webrtc.H264:
		codec = webrtc.NewRTPH264Codec(webrtc.DefaultPayloadTypeH264, 90000)
	case webrtc.Opus:
		codec = webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000)
	case webrtc.G722:
		codec = webrtc.NewRTPG722Codec(webrtc.DefaultPayloadTypeG722, 8000)
	case webrtc.PCMU:
		codec = webrtc.NewRTPPCMUCodec(webrtc.DefaultPayloadTypePCMU, 8000)
	case webrtc.PCMA:
		codec = webrtc.NewRTPPCMACodec(webrtc.DefaultPayloadTypePCMA, 8000)
	default:
		return nil, nil, fmt.Errorf("unknown codec %s", codecName)
	}

	m.engine.RegisterCodec(codec)
	track, err := webrtc.NewTrack(codec.PayloadType, rand.Uint32(), "stream", "stream", codec)
	if err != nil {
		return nil, nil, err
	}

	return pipeline, track, nil
}
