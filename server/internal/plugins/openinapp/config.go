package openinapp

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Enabled bool
	OpenCommand string
}

func (Config) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("openinapp.enabled", false, "whether to enable openinapp plugin")
	if err := viper.BindPFlag("openinapp.enabled", cmd.PersistentFlags().Lookup("openinapp.enabled")); err != nil {
		return err
	}
	cmd.PersistentFlags().String("openinapp.open_command", "xdg-open", "command used to open URLs inside the app")
	if err := viper.BindPFlag("openinapp.open_command", cmd.PersistentFlags().Lookup("openinapp.open_command")); err != nil {
		return err
	}


	return nil
}

func (s *Config) Set() {
	s.Enabled = viper.GetBool("openinapp.enabled")
	s.OpenCommand = viper.GetString("openinapp.open_command")
}
