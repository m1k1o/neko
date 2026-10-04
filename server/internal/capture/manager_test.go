package capture

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/m1k1o/neko/server/internal/config"
)

// captureConfigFromEnv runs the production config flow (init flags, read env,
// populate values) and returns the resulting capture config.
func captureConfigFromEnv(t *testing.T, pipelines string) *config.Capture {
	t.Helper()

	viper.Reset()
	viper.SetEnvPrefix("NEKO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	t.Setenv("NEKO_CAPTURE_VIDEO_PIPELINES", pipelines)

	cfg := &config.Capture{}
	cmd := &cobra.Command{}
	if err := cfg.Init(cmd); err != nil {
		t.Fatal(err)
	}
	cfg.Set()
	return cfg
}

// End to end check for the reported panic: NEKO_CAPTURE_VIDEO_PIPELINES alone
// must leave the capture manager with a usable default video stream.
func TestVideoIDsAvailableWithOnlyPipelinesEnv(t *testing.T) {
	const pipeline = "ximagesrc display-name={display} show-pointer=false ! appsink name=appsink"
	cfg := captureConfigFromEnv(t, `{"high":"`+pipeline+`","medium":"`+pipeline+`","low":"`+pipeline+`"}`)

	manager := New(nil, cfg)

	got := manager.Video().IDs()
	want := []string{"high", "low", "medium"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("capture Video().IDs() = %v, want %v (empty list panicked signalRequest)", got, want)
	}
}
