package chat

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Enabled bool
}

func (Config) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("chat.enabled", true, "whether to enable chat plugin")
	if err := viper.BindPFlag("chat.enabled", cmd.PersistentFlags().Lookup("chat.enabled")); err != nil {
		return err
	}

	return nil
}

func (s *Config) Set() {
	s.Enabled = viper.GetBool("chat.enabled")
}
