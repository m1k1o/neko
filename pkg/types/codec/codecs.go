package codec

import (
	"strings"

	"github.com/pion/webrtc/v3"
)

var RTCPFeedback = []webrtc.RTCPFeedback{
	{Type: webrtc.TypeRTCPFBTransportCC, Parameter: ""},
	{Type: webrtc.TypeRTCPFBGoogREMB, Parameter: ""}, // TODO: Deprecated.

	// https://www.iana.org/assignments/sdp-parameters/sdp-parameters.xhtml#sdp-parameters-19
	{Type: webrtc.TypeRTCPFBCCM, Parameter: "fir"},

	// https://www.iana.org/assignments/sdp-parameters/sdp-parameters.xhtml#sdp-parameters-15
	{Type: webrtc.TypeRTCPFBNACK, Parameter: "pli"},
	{Type: webrtc.TypeRTCPFBNACK, Parameter: ""},
}

func ParseRTC(codec webrtc.RTPCodecParameters) (RTPCodec, bool) {
	codecName := strings.Split(codec.RTPCodecCapability.MimeType, "/")[1]
	return ParseStr(codecName)
}

func ParseStr(codecName string) (codec RTPCodec, ok bool) {
	ok = true

	switch strings.ToLower(codecName) {
	case VP8().Name:
		codec = VP8()
	case VP9().Name:
		codec = VP9()
	case H264().Name:
		codec = H264()
	case Opus().Name:
		codec = Opus()
	case G722().Name:
		codec = G722()
	case PCMU().Name:
		codec = PCMU()
	case PCMA().Name:
		codec = PCMA()
	default:
		ok = false
	}

	return
}

type RTPCodec struct {
	Name        string
	PayloadType webrtc.PayloadType
	Type        webrtc.RTPCodecType
	Capability  webrtc.RTPCodecCapability
	Pipeline    string
}

func (codec *RTPCodec) Register(engine *webrtc.MediaEngine) error {
	return engine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: codec.Capability,
		PayloadType:        codec.PayloadType,
	}, codec.Type)
}

func (codec *RTPCodec) IsVideo() bool {
	return codec.Type == webrtc.RTPCodecTypeVideo
}

func (codec *RTPCodec) IsAudio() bool {
	return codec.Type == webrtc.RTPCodecTypeAudio
}

func (codec *RTPCodec) String() string {
	return codec.Type.String() + "/" + codec.Name
}

func VP8() RTPCodec {
	return RTPCodec{
		Name:        "vp8",
		PayloadType: 96,
		Type:        webrtc.RTPCodecTypeVideo,
		Capability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeVP8,
			ClockRate:    90000,
			Channels:     0,
			SDPFmtpLine:  "",
			RTCPFeedback: RTCPFeedback,
		},
		// https://gstreamer.freedesktop.org/documentation/vpx/vp8enc.html
		// gstreamer1.0-plugins-good
		Pipeline: "vp8enc cpu-used=16 threads=4 deadline=1 error-resilient=partitions keyframe-max-dist=15 static-threshold=20",
	}
}

// TODO: Profile ID.
func VP9() RTPCodec {
	return RTPCodec{
		Name:        "vp9",
		PayloadType: 98,
		Type:        webrtc.RTPCodecTypeVideo,
		Capability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeVP9,
			ClockRate:    90000,
			Channels:     0,
			SDPFmtpLine:  "profile-id=0",
			RTCPFeedback: RTCPFeedback,
		},
		// https://gstreamer.freedesktop.org/documentation/vpx/vp9enc.html
		// gstreamer1.0-plugins-good
		Pipeline: "vp9enc cpu-used=16 threads=4 deadline=1 keyframe-max-dist=15 static-threshold=20",
	}
}

// TODO: Profile ID.
func H264() RTPCodec {
	return RTPCodec{
		Name:        "h264",
		PayloadType: 102,
		Type:        webrtc.RTPCodecTypeVideo,
		Capability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH264,
			ClockRate:    90000,
			Channels:     0,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
			RTCPFeedback: RTCPFeedback,
		},
		// https://gstreamer.freedesktop.org/documentation/x264/index.html
		// gstreamer1.0-plugins-ugly
		Pipeline: "video/x-raw,format=I420 ! x264enc threads=4 bitrate=4096 key-int-max=15 byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream",
		// https://gstreamer.freedesktop.org/documentation/openh264/openh264enc.html
		// gstreamer1.0-plugins-bad
		//Pipeline: "openh264enc multi-thread=4 complexity=high bitrate=3072000 max-bitrate=4096000 ! video/x-h264,stream-format=byte-stream",
	}
}

func Opus() RTPCodec {
	return RTPCodec{
		Name:        "opus",
		PayloadType: 111,
		Type:        webrtc.RTPCodecTypeAudio,
		Capability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeOpus,
			ClockRate:    48000,
			Channels:     2,
			SDPFmtpLine:  "useinbandfec=1;stereo=1",
			RTCPFeedback: []webrtc.RTCPFeedback{},
		},
		// https://gstreamer.freedesktop.org/documentation/opus/opusenc.html
		// gstreamer1.0-plugins-base
		Pipeline: "opusenc inband-fec=true bitrate=128000",
	}
}

func G722() RTPCodec {
	return RTPCodec{
		Name:        "g722",
		PayloadType: 9,
		Type:        webrtc.RTPCodecTypeAudio,
		Capability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeG722,
			ClockRate:    8000,
			Channels:     0,
			SDPFmtpLine:  "",
			RTCPFeedback: []webrtc.RTCPFeedback{},
		},
		// https://gstreamer.freedesktop.org/documentation/libav/avenc_g722.html
		// gstreamer1.0-libav
		Pipeline: "avenc_g722",
	}
}

func PCMU() RTPCodec {
	return RTPCodec{
		Name:        "pcmu",
		PayloadType: 0,
		Type:        webrtc.RTPCodecTypeAudio,
		Capability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypePCMU,
			ClockRate:    8000,
			Channels:     0,
			SDPFmtpLine:  "",
			RTCPFeedback: []webrtc.RTCPFeedback{},
		},
		// https://gstreamer.freedesktop.org/documentation/mulaw/mulawenc.html
		// gstreamer1.0-plugins-good
		Pipeline: "audio/x-raw, rate=8000 ! mulawenc",
	}
}

func PCMA() RTPCodec {
	return RTPCodec{
		Name:        "pcma",
		PayloadType: 8,
		Type:        webrtc.RTPCodecTypeAudio,
		Capability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypePCMA,
			ClockRate:    8000,
			Channels:     0,
			SDPFmtpLine:  "",
			RTCPFeedback: []webrtc.RTCPFeedback{},
		},
		// https://gstreamer.freedesktop.org/documentation/alaw/alawenc.html
		// gstreamer1.0-plugins-good
		Pipeline: "audio/x-raw, rate=8000 ! alawenc",
	}
}
