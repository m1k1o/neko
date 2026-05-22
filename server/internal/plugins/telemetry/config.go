package telemetry

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Enabled  bool
	Endpoint string
}

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

func (c *Config) Set() {
	c.Enabled = viper.GetBool("telemetry.enabled")
	c.Endpoint = viper.GetString("telemetry.endpoint")
}
