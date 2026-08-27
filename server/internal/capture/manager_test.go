package capture

import (
	"strings"
	"testing"

	"github.com/m1k1o/neko/server/internal/config"
)

func TestReplaceCapturePlaceholdersForWindowRegion(t *testing.T) {
	captureConfig := &config.Capture{
		Display:      ":99.0",
		WindowX:      1288,
		WindowY:      728,
		WindowWidth:  1280,
		WindowHeight: 720,
	}
	pipeline, err := replaceCapturePlaceholders(
		"ximagesrc display-name={display} startx={window_x} starty={window_y} "+
			"endx={window_end_x} endy={window_end_y}",
		captureConfig,
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"display-name=:99.0",
		"startx=1288",
		"starty=728",
		"endx=2567",
		"endy=1447",
	} {
		if !strings.Contains(pipeline, expected) {
			t.Fatalf("pipeline %q does not contain %q", pipeline, expected)
		}
	}
}

func TestReplaceCapturePlaceholdersRejectsUnboundedRegionPipeline(t *testing.T) {
	_, err := replaceCapturePlaceholders("ximagesrc display-name={display}", &config.Capture{
		WindowWidth:  1280,
		WindowHeight: 720,
	})
	if err == nil {
		t.Fatal("unbounded capture pipeline was accepted for a window region")
	}
}

func TestXImageSourceCropsWindowRegion(t *testing.T) {
	source := xImageSource(&config.Capture{
		Display:      ":99.0",
		WindowX:      8,
		WindowY:      16,
		WindowWidth:  1280,
		WindowHeight: 720,
	}, true)
	for _, expected := range []string{"startx=8", "starty=16", "endx=1287", "endy=735"} {
		if !strings.Contains(source, expected) {
			t.Fatalf("source %q does not contain %q", source, expected)
		}
	}
}
