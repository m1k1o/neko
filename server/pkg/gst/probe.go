package gst

import "fmt"

// ProbeEncoder creates a short-lived live GStreamer pipeline for a hardware
// encoder and waits for it to reach PLAYING. A factory can be present in the
// registry while its VAAPI/NVENC device or driver is unusable, so checking the
// factory alone is not enough to select an encoder safely.
//
// The probe intentionally uses videotestsrc instead of the desktop capture
// source. It therefore does not require an X display and only tests the
// encoder, parser, and sink path that a Chromium profile needs.
func ProbeEncoder(element string) error {
	if err := CheckElement(element); err != nil {
		return err
	}

	pipelineSpec := fmt.Sprintf(
		"videotestsrc is-live=true pattern=black ! video/x-raw,format=NV12,width=1280,height=720,framerate=30/1 ! %s ! h264parse ! fakesink sync=false",
		element,
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
