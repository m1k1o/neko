package codec

import "github.com/pion/webrtc/v3"

type RTPCodec struct {
	Name        string
	PayloadType webrtc.PayloadType
	Type        webrtc.RTPCodecType
	Capability  webrtc.RTPCodecCapability
}

func (codec *RTPCodec) Register(engine *webrtc.MediaEngine) error {
	return engine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: codec.Capability,
		PayloadType:        codec.PayloadType,
	}, codec.Type)
}

func VP8() RTPCodec {
	return RTPCodec{
		Name: "vp8",
		PayloadType: 96,
		Type: webrtc.RTPCodecTypeVideo,
		Capability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeVP8,
			ClockRate: 90000,
			Channels: 0,
			SDPFmtpLine: "",
			RTCPFeedback: nil,
		},
	}
}

// TODO: Profile ID.
func VP9() RTPCodec {
	return RTPCodec{
		Name: "vp9",
		PayloadType: 98,
		Type: webrtc.RTPCodecTypeVideo,
		Capability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeVP9,
			ClockRate: 90000,
			Channels: 0,
			SDPFmtpLine: "profile-id=0",
			RTCPFeedback: nil,
		},
	}
}

// TODO: Profile ID.
func H264() RTPCodec {
	return RTPCodec{
		Name: "h264",
		PayloadType: 102,
		Type: webrtc.RTPCodecTypeVideo,
		Capability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264,
			ClockRate: 90000,
			Channels: 0,
			SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
			RTCPFeedback: nil,
		},
	}
}

func Opus() RTPCodec {
	return RTPCodec{
		Name: "opus",
		PayloadType: 111,
		Type: webrtc.RTPCodecTypeAudio,
		Capability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeOpus,
			ClockRate: 48000,
			Channels: 2,
			SDPFmtpLine: "",
			RTCPFeedback: nil,
		},
	}
}

func G722() RTPCodec {
	return RTPCodec{
		Name: "g722",
		PayloadType: 9,
		Type: webrtc.RTPCodecTypeAudio,
		Capability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeG722,
			ClockRate: 8000,
			Channels: 0,
			SDPFmtpLine: "",
			RTCPFeedback: nil,
		},
	}
}

func PCMU() RTPCodec {
	return RTPCodec{
		Name: "pcmu",
		PayloadType: 0,
		Type: webrtc.RTPCodecTypeAudio,
		Capability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypePCMU,
			ClockRate: 8000,
			Channels: 0,
			SDPFmtpLine: "",
			RTCPFeedback: nil,
		},
	}
}

func PCMA() RTPCodec {
	return RTPCodec{
		Name: "pcma",
		PayloadType: 8,
		Type: webrtc.RTPCodecTypeAudio,
		Capability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypePCMA,
			ClockRate: 8000,
			Channels: 0,
			SDPFmtpLine: "",
			RTCPFeedback: nil,
		},
	}
}
