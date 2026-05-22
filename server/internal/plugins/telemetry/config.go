package telemetry

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config controls the telemetry-forwarding plugin which mirrors WebRTC
// session connect/disconnect events into the kernel-images-api telemetry
// stream as live_view_connect / live_view_disconnect events.
type Config struct {
	Enabled  bool
	Endpoint string
}

// Init registers CLI/viper flags. Defaults are wired for the in-VM headful
// image where neko and kernel-images-api share localhost.
func (Config) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("telemetry.enabled", false, "forward live-view session connect/disconnect events to kernel-images-api")
	if err := viper.BindPFlag("telemetry.enabled", cmd.PersistentFlags().Lookup("telemetry.enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("telemetry.endpoint", "http://127.0.0.1:10001/telemetry/events", "kernel-images-api telemetry publish endpoint")
	if err := viper.BindPFlag("telemetry.endpoint", cmd.PersistentFlags().Lookup("telemetry.endpoint")); err != nil {
		return err
	}

	return nil
}

// Set hydrates Config from viper after flag/file/env parsing.
func (c *Config) Set() {
	c.Enabled = viper.GetBool("telemetry.enabled")
	c.Endpoint = viper.GetString("telemetry.endpoint")
}
