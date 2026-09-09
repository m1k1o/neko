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

// VideoConfig generates the standard software-encoding pipeline description
// for a profile. Hardware encoder selection is handled by the next M1
// capability-discovery increment.
func (p Profile) VideoConfig(rtpCodec codec.RTPCodec, showPointer bool) (types.VideoConfig, error) {
	config := types.VideoConfig{
		Width:       strconv.Itoa(p.Width),
		Height:      strconv.Itoa(p.Height),
		Fps:         strconv.Itoa(p.FPS),
		Bitrate:     p.BitrateKbps,
		ShowPointer: showPointer,
	}

	switch rtpCodec.Name {
	case codec.VP8().Name:
		config.GstEncoder = "vp8enc"
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
		config.GstPrefix = "! video/x-raw,format=I420"
		config.GstEncoder = "x264enc"
		config.GstParams = map[string]string{
			"threads":      "4",
			"bitrate":      strconv.Itoa(p.BitrateKbps),
			"key-int-max":  strconv.Itoa(p.FPS),
			"byte-stream":  "true",
			"tune":         "zerolatency",
			"speed-preset": "veryfast",
			"bframes":      "0",
		}
		config.GstSuffix = "! video/x-h264,stream-format=byte-stream,profile=constrained-baseline"
	default:
		return types.VideoConfig{}, fmt.Errorf("quality profiles support only vp8 or h264, got %s", rtpCodec.Name)
	}

	return config, nil
}
