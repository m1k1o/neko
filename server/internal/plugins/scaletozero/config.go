package scaletozero

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Enabled bool
}

func (Config) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("scaletozero.enabled", false, "enable scale-to-zero")
	if err := viper.BindPFlag("scaletozero.enabled", cmd.PersistentFlags().Lookup("scaletozero.enabled")); err != nil {
		return err
	}

	return nil
}

func (c *Config) Set() {
	c.Enabled = viper.GetBool("scaletozero.enabled")
}
