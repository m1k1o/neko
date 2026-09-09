package gst

import (
	"strings"
	"testing"

	"github.com/m1k1o/neko/server/pkg/types/codec"
)

func TestProbeEncoderRejectsMissingElement(t *testing.T) {
	err := ProbeEncoder(codec.H264(), "neko-missing-h264-encoder")
	if err == nil || !strings.Contains(err.Error(), "required gstreamer element") {
		t.Fatalf("unexpected missing encoder probe result: %v", err)
	}
}

func TestProbeEncoderRejectsUnknownCodec(t *testing.T) {
	err := ProbeEncoder(codec.Opus(), "neko-missing-audio-encoder")
	if err == nil || !strings.Contains(err.Error(), "unsupported encoder probe codec") {
		t.Fatalf("unexpected unknown codec probe result: %v", err)
	}
}
