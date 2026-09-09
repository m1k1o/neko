package quality

import (
	"fmt"
	"strings"

	"github.com/m1k1o/neko/server/pkg/types/codec"
)

type Encoder string

const (
	EncoderAuto     Encoder = "auto"
	EncoderSoftware Encoder = "software"
	EncoderVAAPI    Encoder = "vaapi"
	EncoderNVENC    Encoder = "nvenc"
)

type ElementProbe func(element string) error

type EncoderSelection struct {
	Requested   Encoder
	Encoder     Encoder
	Codec       codec.RTPCodec
	Element     string
	Unavailable []string
}

func ParseEncoder(value string) (Encoder, error) {
	encoder := Encoder(strings.ToLower(strings.TrimSpace(value)))
	if encoder == "" {
		return EncoderAuto, nil
	}
	switch encoder {
	case EncoderAuto, EncoderSoftware, EncoderVAAPI, EncoderNVENC:
		return encoder, nil
	default:
		return "", fmt.Errorf("unsupported video encoder %q (want auto, software, vaapi, or nvenc)", value)
	}
}

// ResolveEncoder selects an encoder whose complete GStreamer element chain is
// available. Hardware requests deliberately fall back to software and then
// VP8 so a missing device or plugin is diagnosed at startup instead of
// producing an unusable video stream.
func ResolveEncoder(preferred codec.RTPCodec, requested Encoder, probe ElementProbe) (EncoderSelection, error) {
	if probe == nil {
		return EncoderSelection{}, fmt.Errorf("video encoder element probe is required")
	}

	if preferred.Name == codec.VP8().Name {
		if requested != EncoderAuto && requested != EncoderSoftware {
			return EncoderSelection{}, fmt.Errorf("%s cannot encode vp8 in the Chromium M1 quality profiles", requested)
		}
		return resolveCandidates(requested, probe, []encoderCandidate{vp8Candidate()})
	}
	if preferred.Name != codec.H264().Name {
		return EncoderSelection{}, fmt.Errorf("encoder selection supports only vp8 or h264, got %s", preferred.Name)
	}

	var candidates []encoderCandidate
	switch requested {
	case EncoderAuto:
		candidates = []encoderCandidate{nvencCandidate(), vaapiCandidate(), x264Candidate(), vp8Candidate()}
	case EncoderNVENC:
		candidates = []encoderCandidate{nvencCandidate(), x264Candidate(), vp8Candidate()}
	case EncoderVAAPI:
		candidates = []encoderCandidate{vaapiCandidate(), x264Candidate(), vp8Candidate()}
	case EncoderSoftware:
		candidates = []encoderCandidate{x264Candidate(), vp8Candidate()}
	default:
		return EncoderSelection{}, fmt.Errorf("unsupported video encoder %q", requested)
	}

	return resolveCandidates(requested, probe, candidates)
}

type encoderCandidate struct {
	encoder      Encoder
	codec        codec.RTPCodec
	alternatives []string
	required     []string
}

func nvencCandidate() encoderCandidate {
	return encoderCandidate{
		encoder:      EncoderNVENC,
		codec:        codec.H264(),
		alternatives: []string{"nvautogpuh264enc", "nvh264enc"},
		required:     []string{"h264parse"},
	}
}

func vaapiCandidate() encoderCandidate {
	return encoderCandidate{
		encoder:      EncoderVAAPI,
		codec:        codec.H264(),
		alternatives: []string{"vah264enc"},
		required:     []string{"h264parse"},
	}
}

func x264Candidate() encoderCandidate {
	return encoderCandidate{
		encoder:      EncoderSoftware,
		codec:        codec.H264(),
		alternatives: []string{"x264enc"},
		required:     []string{"h264parse"},
	}
}

func vp8Candidate() encoderCandidate {
	return encoderCandidate{
		encoder:      EncoderSoftware,
		codec:        codec.VP8(),
		alternatives: []string{"vp8enc"},
	}
}

func resolveCandidates(requested Encoder, probe ElementProbe, candidates []encoderCandidate) (EncoderSelection, error) {
	unavailable := make([]string, 0)
	for _, candidate := range candidates {
		element := ""
		for _, alternative := range candidate.alternatives {
			if err := probe(alternative); err != nil {
				unavailable = append(unavailable, alternative)
				continue
			}
			element = alternative
			break
		}
		if element == "" {
			continue
		}

		complete := true
		for _, required := range candidate.required {
			if err := probe(required); err != nil {
				unavailable = append(unavailable, required)
				complete = false
				break
			}
		}
		if !complete {
			continue
		}

		return EncoderSelection{
			Requested:   requested,
			Encoder:     candidate.encoder,
			Codec:       candidate.codec,
			Element:     element,
			Unavailable: unavailable,
		}, nil
	}

	return EncoderSelection{}, fmt.Errorf("no usable Chromium M1 video encoder found (missing: %s)", strings.Join(unavailable, ", "))
}
