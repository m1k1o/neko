// Package quality defines conservative, explicit video quality profiles for
// the Chromium M1 deployment. Profiles are opt-in so existing pipelines keep
// their historical behavior.
package quality

import (
	"fmt"
	"strconv"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

type Name string

const (
	Low      Name = "low"
	Balanced Name = "balanced"
	High     Name = "high"
)

type Profile struct {
	Name        Name
	Width       int
	Height      int
	FPS         int
	BitrateKbps int
}

func Parse(value string) (Profile, error) {
	switch Name(value) {
	case Low:
		return Profile{Name: Low, Width: 854, Height: 480, FPS: 20, BitrateKbps: 1000}, nil
	case Balanced:
		return Profile{Name: Balanced, Width: 1280, Height: 720, FPS: 30, BitrateKbps: 2500}, nil
	case High:
		return Profile{Name: High, Width: 1920, Height: 1080, FPS: 30, BitrateKbps: 4500}, nil
	default:
		return Profile{}, fmt.Errorf("unsupported video quality profile %q (want low, balanced, or high)", value)
	}
}

// VideoConfig generates the standard pipeline description for an encoder
// selection that has already passed capability discovery.
func (p Profile) VideoConfig(rtpCodec codec.RTPCodec, encoder Encoder, element string, showPointer bool) (types.VideoConfig, error) {
	config := types.VideoConfig{
		Width:       strconv.Itoa(p.Width),
		Height:      strconv.Itoa(p.Height),
		Fps:         strconv.Itoa(p.FPS),
		Bitrate:     p.BitrateKbps,
		ShowPointer: showPointer,
	}

	switch rtpCodec.Name {
	case codec.VP8().Name:
		if encoder != EncoderSoftware || element != "vp8enc" {
			return types.VideoConfig{}, fmt.Errorf("invalid vp8 encoder selection %s/%s", encoder, element)
		}
		config.GstEncoder = element
		config.GstParams = map[string]string{
			// round keeps gval's result integral. A bare numeric literal is
			// evaluated as float64 and may be rendered in scientific notation,
			// which is not accepted consistently across GStreamer versions.
			"target-bitrate":    fmt.Sprintf("round(%d * 1000)", p.BitrateKbps),
			"cpu-used":          "4",
			"end-usage":         "cbr",
			"threads":           "4",
			"deadline":          "1",
			"keyframe-max-dist": strconv.Itoa(p.FPS),
		}
	case codec.H264().Name:
		config.GstPrefix = "! video/x-raw,format=NV12"
		config.GstEncoder = element
		config.GstSuffix = "! h264parse config-interval=-1 ! video/x-h264,stream-format=byte-stream,profile=constrained-baseline"
		switch encoder {
		case EncoderSoftware:
			if element != "x264enc" {
				return types.VideoConfig{}, fmt.Errorf("invalid software h264 encoder %s", element)
			}
			config.GstPrefix = "! video/x-raw,format=I420"
			config.GstParams = map[string]string{
				"threads":      "4",
				"bitrate":      strconv.Itoa(p.BitrateKbps),
				"key-int-max":  strconv.Itoa(p.FPS),
				"byte-stream":  "true",
				"tune":         "zerolatency",
				"speed-preset": "veryfast",
				"bframes":      "0",
			}
		case EncoderVAAPI:
			if element != "vah264enc" && element != "vah264lpenc" {
				return types.VideoConfig{}, fmt.Errorf("invalid VAAPI h264 encoder %s", element)
			}
			config.GstParams = map[string]string{
				"rate-control": "cbr",
				"bitrate":      strconv.Itoa(p.BitrateKbps),
				"key-int-max":  strconv.Itoa(p.FPS),
				"target-usage": "7",
			}
		case EncoderNVENC:
			if element != "nvautogpuh264enc" && element != "nvh264enc" {
				return types.VideoConfig{}, fmt.Errorf("invalid NVENC h264 encoder %s", element)
			}
			config.GstParams = map[string]string{
				"preset":          "2",
				"gop-size":        strconv.Itoa(p.FPS),
				"spatial-aq":      "true",
				"temporal-aq":     "true",
				"bitrate":         strconv.Itoa(p.BitrateKbps),
				"vbv-buffer-size": strconv.Itoa(p.BitrateKbps),
				"rc-mode":         "6",
			}
		default:
			return types.VideoConfig{}, fmt.Errorf("invalid h264 encoder selection %s", encoder)
		}
	case codec.H265().Name:
		config.GstPrefix = "! video/x-raw,format=NV12"
		config.GstEncoder = element
		config.GstSuffix = "! h265parse config-interval=-1 ! video/x-h265,stream-format=byte-stream,alignment=au,profile=main"
		switch encoder {
		case EncoderSoftware:
			if element != "x265enc" {
				return types.VideoConfig{}, fmt.Errorf("invalid software h265 encoder %s", element)
			}
			config.GstPrefix = "! video/x-raw,format=I420"
			config.GstParams = map[string]string{
				"bitrate":      strconv.Itoa(p.BitrateKbps),
				"key-int-max":  strconv.Itoa(p.FPS),
				"tune":         "zerolatency",
				"speed-preset": "veryfast",
			}
		case EncoderVAAPI:
			if element != "vah265enc" && element != "vah265lpenc" {
				return types.VideoConfig{}, fmt.Errorf("invalid VAAPI h265 encoder %s", element)
			}
			config.GstParams = map[string]string{
				"rate-control": "cbr",
				"bitrate":      strconv.Itoa(p.BitrateKbps),
				"key-int-max":  strconv.Itoa(p.FPS),
				"target-usage": "7",
			}
		case EncoderNVENC:
			if element != "nvautogpuh265enc" && element != "nvh265enc" {
				return types.VideoConfig{}, fmt.Errorf("invalid NVENC h265 encoder %s", element)
			}
			config.GstParams = map[string]string{
				"preset":          "2",
				"gop-size":        strconv.Itoa(p.FPS),
				"spatial-aq":      "true",
				"temporal-aq":     "true",
				"bitrate":         strconv.Itoa(p.BitrateKbps),
				"vbv-buffer-size": strconv.Itoa(p.BitrateKbps),
				"rc-mode":         "6",
			}
		default:
			return types.VideoConfig{}, fmt.Errorf("invalid h265 encoder selection %s", encoder)
		}
	case codec.AV1().Name:
		config.GstPrefix = "! video/x-raw,format=NV12"
		config.GstEncoder = element
		config.GstSuffix = "! video/x-av1,stream-format=obu-stream,alignment=tu"
		switch encoder {
		case EncoderSoftware:
			config.GstPrefix = "! video/x-raw,format=I420"
			switch element {
			case "av1enc":
				config.GstParams = map[string]string{
					"target-bitrate":    strconv.Itoa(p.BitrateKbps),
					"cpu-used":          "8",
					"end-usage":         "cbr",
					"usage-profile":     "realtime",
					"threads":           "4",
					"lag-in-frames":     "0",
					"keyframe-max-dist": strconv.Itoa(p.FPS),
				}
			case "svtav1enc":
				config.GstParams = map[string]string{
					"target-bitrate":      strconv.Itoa(p.BitrateKbps),
					"preset":              "10",
					"intra-period-length": strconv.Itoa(p.FPS),
				}
			default:
				return types.VideoConfig{}, fmt.Errorf("invalid software av1 encoder %s", element)
			}
		case EncoderVAAPI:
			if element != "vaav1enc" {
				return types.VideoConfig{}, fmt.Errorf("invalid VAAPI av1 encoder %s", element)
			}
			config.GstParams = map[string]string{
				"rate-control": "cbr",
				"bitrate":      strconv.Itoa(p.BitrateKbps),
				"key-int-max":  strconv.Itoa(p.FPS),
				"target-usage": "7",
			}
		case EncoderNVENC:
			if element != "nvautogpuav1enc" && element != "nvav1enc" {
				return types.VideoConfig{}, fmt.Errorf("invalid NVENC av1 encoder %s", element)
			}
			config.GstParams = map[string]string{
				"preset":          "2",
				"gop-size":        strconv.Itoa(p.FPS),
				"spatial-aq":      "true",
				"temporal-aq":     "true",
				"bitrate":         strconv.Itoa(p.BitrateKbps),
				"vbv-buffer-size": strconv.Itoa(p.BitrateKbps),
				"rc-mode":         "6",
			}
		default:
			return types.VideoConfig{}, fmt.Errorf("invalid av1 encoder selection %s", encoder)
		}
	default:
		return types.VideoConfig{}, fmt.Errorf("quality profiles support only vp8, av1, h264, or h265, got %s", rtpCodec.Name)
	}

	return config, nil
}
