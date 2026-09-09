package gst

import (
	"strings"
	"testing"
)

func TestProbeEncoderRejectsMissingElement(t *testing.T) {
	err := ProbeEncoder("neko-missing-h264-encoder")
	if err == nil || !strings.Contains(err.Error(), "required gstreamer element") {
		t.Fatalf("unexpected missing encoder probe result: %v", err)
	}
}
