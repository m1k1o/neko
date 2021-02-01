package codec

import "github.com/pion/webrtc/v3"

const (
	VP8 = "vp8"
	VP9 = "vp9"
	H264 = "h264"
	Opus = "opus"
	G722 = "g722"
	PCMU = "pcmu"
	PCMA = "pcma"
)

type RTP struct {
	Name        string
	PayloadType webrtc.PayloadType
	Type        webrtc.RTPCodecType
	Capability  webrtc.RTPCodecCapability
}

func New(codecType string) RTP {
	codec := RTP{}

	switch codecType {
	case "vp8":
		codec.Name = "vp8"
		codec.PayloadType = 96
		codec.Type = webrtc.RTPCodecTypeVideo
		codec.Capability = webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeVP8,
			ClockRate: 90000,
			Channels: 0,
			SDPFmtpLine: "",
			RTCPFeedback: nil,
		}
	case "vp9":
		codec.Name = "vp9"
		codec.PayloadType = 98
		codec.Type = webrtc.RTPCodecTypeVideo
		codec.Capability = webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeVP9,
			ClockRate: 90000,
			Channels: 0,
			SDPFmtpLine: "profile-id=0",
			RTCPFeedback: nil,
		}
	case "h264":
		codec.Name = "h264"
		codec.PayloadType = 102
		codec.Type = webrtc.RTPCodecTypeVideo
		codec.Capability = webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264,
			ClockRate: 90000,
			Channels: 0,
			SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
			RTCPFeedback: nil,
		}
	case "opus":
		codec.Name = "opus"
		codec.PayloadType = 111
		codec.Type = webrtc.RTPCodecTypeAudio
		codec.Capability = webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeOpus,
			ClockRate: 48000,
			Channels: 2,
			SDPFmtpLine: "",
			RTCPFeedback: nil,
		}
	case "g722":
		codec.Name = "g722"
		codec.PayloadType = 9
		codec.Type = webrtc.RTPCodecTypeAudio
		codec.Capability = webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeG722,
			ClockRate: 8000,
			Channels: 0,
			SDPFmtpLine: "",
			RTCPFeedback: nil,
		}
	case "pcmu":
		codec.Name = "pcmu"
		codec.PayloadType = 0
		codec.Type = webrtc.RTPCodecTypeAudio
		codec.Capability = webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypePCMU,
			ClockRate: 8000,
			Channels: 0,
			SDPFmtpLine: "",
			RTCPFeedback: nil,
		}
	case "pcma":
		codec.Name = "pcma"
		codec.PayloadType = 8
		codec.Type = webrtc.RTPCodecTypeAudio
		codec.Capability = webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypePCMA,
			ClockRate: 8000,
			Channels: 0,
			SDPFmtpLine: "",
			RTCPFeedback: nil,
		}
	}

	return codec
}

func (codec *RTP) Register(engine *webrtc.MediaEngine) error {
	return engine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: codec.Capability,
		PayloadType:        codec.PayloadType,
	}, codec.Type)
}
