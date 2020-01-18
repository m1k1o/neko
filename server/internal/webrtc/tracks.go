package webrtc

import (
	"fmt"
	"math/rand"

	"github.com/pion/webrtc/v2"
	"github.com/pkg/errors"

	"n.eko.moe/neko/internal/gst"
)

func (m *WebRTCManager) createVideoTrack(payloadType uint8) error {

	clockrate := uint32(90000)
	var codec *webrtc.RTPCodec
	switch payloadType {
	case webrtc.DefaultPayloadTypeVP8:
		codec = webrtc.NewRTPVP8Codec(payloadType, clockrate)
		break
	case webrtc.DefaultPayloadTypeVP9:
		codec = webrtc.NewRTPVP9Codec(payloadType, clockrate)
		break
	case webrtc.DefaultPayloadTypeH264:
		codec = webrtc.NewRTPH264Codec(payloadType, clockrate)
		break
	default:
		return errors.Errorf("unknown video codec %s", payloadType)
	}

	track, err := webrtc.NewTrack(payloadType, rand.Uint32(), "stream", "stream", codec)
	if err != nil {
		return err
	}

	var pipeline *gst.Pipeline
	src := fmt.Sprintf("ximagesrc xid=%s show-pointer=true use-damage=false ! video/x-raw,framerate=30/1 ! videoconvert ! queue", m.conf.Display)
	switch payloadType {
	case webrtc.DefaultPayloadTypeVP8:
		pipeline = gst.CreatePipeline(webrtc.VP8, []*webrtc.Track{track}, src)
		break
	case webrtc.DefaultPayloadTypeVP9:
		pipeline = gst.CreatePipeline(webrtc.VP9, []*webrtc.Track{track}, src)
		break
	case webrtc.DefaultPayloadTypeH264:
		pipeline = gst.CreatePipeline(webrtc.H264, []*webrtc.Track{track}, src)
		break
	}

	m.video = track
	m.videoPipeline = pipeline
	return nil
}

func (m *WebRTCManager) createAudioTrack(payloadType uint8) error {
	var codec *webrtc.RTPCodec
	switch payloadType {
	case webrtc.DefaultPayloadTypeOpus:
		codec = webrtc.NewRTPOpusCodec(payloadType, 48000)
		break
	case webrtc.DefaultPayloadTypeG722:
		codec = webrtc.NewRTPG722Codec(payloadType, 48000)
		break
	case webrtc.DefaultPayloadTypePCMU:
		codec = webrtc.NewRTPPCMUCodec(payloadType, 8000)
		break
	case webrtc.DefaultPayloadTypePCMA:
		codec = webrtc.NewRTPPCMACodec(payloadType, 8000)
		break
	default:
		return errors.Errorf("unknown audio codec %s", payloadType)
	}

	track, err := webrtc.NewTrack(payloadType, rand.Uint32(), "stream", "stream", codec)
	if err != nil {
		return err
	}

	var pipeline *gst.Pipeline
	src := fmt.Sprintf("pulsesrc device=%s ! audioconvert", m.conf.Device)
	switch payloadType {
	case webrtc.DefaultPayloadTypeOpus:
		pipeline = gst.CreatePipeline(webrtc.Opus, []*webrtc.Track{track}, src)
		break
	case webrtc.DefaultPayloadTypeG722:
		pipeline = gst.CreatePipeline(webrtc.G722, []*webrtc.Track{track}, src)
		break
	case webrtc.DefaultPayloadTypePCMU:
		pipeline = gst.CreatePipeline(webrtc.PCMU, []*webrtc.Track{track}, src)
		break
	case webrtc.DefaultPayloadTypePCMA:
		pipeline = gst.CreatePipeline(webrtc.PCMA, []*webrtc.Track{track}, src)
		break
	}

	m.audio = track
	m.audioPipeline = pipeline
	return nil
}
