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

// RuntimeProbe validates that an encoder can initialize its runtime device
// and driver. Element discovery alone is insufficient for hardware encoders:
// a GStreamer factory can be registered while VAAPI or NVENC cannot open a
// usable device inside the container.
type RuntimeProbe func(codec codec.RTPCodec, encoder Encoder, element string) error

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
// available. Hardware requests deliberately fall back to software and then a
// broadly supported codec so a missing device or plugin is diagnosed at
// startup instead of producing an unusable video stream.
func ResolveEncoder(preferred codec.RTPCodec, requested Encoder, probe ElementProbe) (EncoderSelection, error) {
	return ResolveEncoderWithRuntime(preferred, requested, probe, nil)
}

// ResolveEncoderWithRuntime selects an encoder after checking both the
// GStreamer factories and, for hardware candidates, their runtime device.
// Failed hardware probes are treated as unavailable so the next same-codec
// software candidate can be selected without waiting for a viewer to join.
// AV1, H.264, and H.265 are tried in their requested codec first; if that
// codec cannot be produced, the resolver falls back to software H.264 and
// finally VP8 for Chromium interoperability.
func ResolveEncoderWithRuntime(preferred codec.RTPCodec, requested Encoder, probe ElementProbe, runtimeProbe RuntimeProbe) (EncoderSelection, error) {
	if probe == nil {
		return EncoderSelection{}, fmt.Errorf("video encoder element probe is required")
	}

	candidates, err := candidatesFor(preferred, requested)
	if err != nil {
		return EncoderSelection{}, err
	}
	return resolveCandidates(requested, probe, runtimeProbe, candidates)
}

type encoderCandidate struct {
	encoder      Encoder
	codec        codec.RTPCodec
	alternatives []string
	required     []string
	hardware     bool
}

func candidatesFor(preferred codec.RTPCodec, requested Encoder) ([]encoderCandidate, error) {
	if preferred.Name == codec.VP8().Name {
		if requested != EncoderAuto && requested != EncoderSoftware {
			return nil, fmt.Errorf("%s cannot encode vp8 in the Chromium M1 quality profiles", requested)
		}
		return []encoderCandidate{vp8Candidate()}, nil
	}

	if preferred.Name != codec.AV1().Name && preferred.Name != codec.H264().Name && preferred.Name != codec.H265().Name {
		return nil, fmt.Errorf("encoder selection supports only vp8, av1, h264, or h265, got %s", preferred.Name)
	}

	var candidates []encoderCandidate
	switch requested {
	case EncoderAuto:
		candidates = append(candidates, nvencCandidate(preferred), vaapiCandidate(preferred))
		candidates = append(candidates, softwareCandidates(preferred)...)
	case EncoderNVENC:
		candidates = append(candidates, nvencCandidate(preferred))
		candidates = append(candidates, softwareCandidates(preferred)...)
	case EncoderVAAPI:
		candidates = append(candidates, vaapiCandidate(preferred))
		candidates = append(candidates, softwareCandidates(preferred)...)
	case EncoderSoftware:
		candidates = append(candidates, softwareCandidates(preferred)...)
	default:
		return nil, fmt.Errorf("unsupported video encoder %q", requested)
	}

	// H.264 is the compatibility bridge for codecs that are newer or less
	// consistently exposed by WebRTC implementations. Keep this fallback
	// software-only so a requested hardware family never silently switches
	// to a different GPU API.
	if preferred.Name != codec.H264().Name {
		candidates = append(candidates, softwareCandidates(codec.H264())...)
	}
	if preferred.Name != codec.VP8().Name {
		candidates = append(candidates, vp8Candidate())
	}
	return candidates, nil
}

func nvencCandidate(videoCodec codec.RTPCodec) encoderCandidate {
	var alternatives []string
	switch videoCodec.Name {
	case codec.AV1().Name:
		alternatives = []string{"nvautogpuav1enc", "nvav1enc"}
	case codec.H264().Name:
		alternatives = []string{"nvautogpuh264enc", "nvh264enc"}
	case codec.H265().Name:
		alternatives = []string{"nvautogpuh265enc", "nvh265enc"}
	default:
		return encoderCandidate{}
	}
	return encoderCandidate{
		encoder:      EncoderNVENC,
		codec:        videoCodec,
		alternatives: alternatives,
		required:     requiredElements(videoCodec),
		hardware:     true,
	}
}

func vaapiCandidate(videoCodec codec.RTPCodec) encoderCandidate {
	var alternatives []string
	switch videoCodec.Name {
	case codec.AV1().Name:
		alternatives = []string{"vaav1enc"}
	case codec.H264().Name:
		alternatives = []string{"vah264enc", "vah264lpenc"}
	case codec.H265().Name:
		alternatives = []string{"vah265enc", "vah265lpenc"}
	default:
		return encoderCandidate{}
	}
	return encoderCandidate{
		encoder:      EncoderVAAPI,
		codec:        videoCodec,
		alternatives: alternatives,
		required:     requiredElements(videoCodec),
		hardware:     true,
	}
}

func softwareCandidates(videoCodec codec.RTPCodec) []encoderCandidate {
	switch videoCodec.Name {
	case codec.AV1().Name:
		return []encoderCandidate{{
			encoder:      EncoderSoftware,
			codec:        videoCodec,
			alternatives: []string{"av1enc", "svtav1enc"},
		}}
	case codec.H264().Name:
		return []encoderCandidate{{
			encoder:      EncoderSoftware,
			codec:        videoCodec,
			alternatives: []string{"x264enc"},
			required:     requiredElements(videoCodec),
		}}
	case codec.H265().Name:
		return []encoderCandidate{{
			encoder:      EncoderSoftware,
			codec:        videoCodec,
			alternatives: []string{"x265enc"},
			required:     requiredElements(videoCodec),
		}}
	default:
		return nil
	}
}

func vp8Candidate() encoderCandidate {
	return encoderCandidate{
		encoder:      EncoderSoftware,
		codec:        codec.VP8(),
		alternatives: []string{"vp8enc"},
	}
}

func requiredElements(videoCodec codec.RTPCodec) []string {
	switch videoCodec.Name {
	case codec.H264().Name:
		return []string{"h264parse"}
	case codec.H265().Name:
		return []string{"h265parse"}
	default:
		// AV1 encoders already produce the OBU stream consumed by the RTP
		// packetizer; a parser is optional and is therefore not required for
		// capability selection.
		return nil
	}
}

func resolveCandidates(requested Encoder, probe ElementProbe, runtimeProbe RuntimeProbe, candidates []encoderCandidate) (EncoderSelection, error) {
	unavailable := make([]string, 0)
	for _, candidate := range candidates {
		for _, alternative := range candidate.alternatives {
			if err := probe(alternative); err != nil {
				unavailable = append(unavailable, alternative)
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

			if candidate.hardware && runtimeProbe != nil {
				if err := runtimeProbe(candidate.codec, candidate.encoder, alternative); err != nil {
					unavailable = append(unavailable, fmt.Sprintf("%s (%v)", alternative, err))
					continue
				}
			}

			return EncoderSelection{
				Requested:   requested,
				Encoder:     candidate.encoder,
				Codec:       candidate.codec,
				Element:     alternative,
				Unavailable: unavailable,
			}, nil
		}
	}

	return EncoderSelection{}, fmt.Errorf("no usable Chromium M1 video encoder found (missing: %s)", strings.Join(unavailable, ", "))
}
