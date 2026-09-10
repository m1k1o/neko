package gst

import (
	"fmt"

	"github.com/m1k1o/neko/server/pkg/types/codec"
)

// ProbeEncoder creates a short-lived live GStreamer pipeline for an encoder
// and waits for it to reach PLAYING. A factory can be present in the registry
// while its VAAPI/NVENC device or driver is unusable, so checking the factory
// alone is not enough to select an encoder safely.
//
// The probe intentionally uses videotestsrc instead of the desktop capture
// source. It therefore does not require an X display and only tests the
// encoder, parser, and sink path that a Chromium profile needs.
func ProbeEncoder(videoCodec codec.RTPCodec, element string) error {
	input := "videotestsrc is-live=true pattern=black ! videoconvert ! video/x-raw,width=1280,height=720,framerate=30/1"
	output := ""
	switch videoCodec.Name {
	case codec.H264().Name:
		output = "! h264parse ! video/x-h264,stream-format=byte-stream"
	case codec.H265().Name:
		output = "! h265parse ! video/x-h265,stream-format=byte-stream"
	case codec.AV1().Name:
		output = "! video/x-av1,stream-format=obu-stream,alignment=tu"
	case codec.VP8().Name:
		// VP8 is currently software-only in the M1 profile, but keeping the
		// generic probe codec-aware prevents accidental parser assumptions.
	default:
		return fmt.Errorf("unsupported encoder probe codec %s", videoCodec.Name)
	}

	if err := CheckElement(element); err != nil {
		return err
	}

	pipelineSpec := fmt.Sprintf(
		"%s ! %s %s ! fakesink sync=false",
		input, element, output,
	)
	pipeline, err := CreatePipeline(pipelineSpec)
	if err != nil {
		return fmt.Errorf("%s pipeline could not be created: %w", element, err)
	}

	if pipeline.Play() {
		pipeline.Destroy()
		return nil
	}

	pipeline.Destroy()
	return fmt.Errorf("%s pipeline failed to reach PLAYING (GPU device or driver may be unavailable)", element)
}
