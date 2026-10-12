package chat

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Enabled bool
	// MaxLength is the longest accepted message in characters (not bytes); 0 disables the limit.
	MaxLength int
}

func (Config) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("chat.enabled", true, "whether to enable chat plugin")
	if err := viper.BindPFlag("chat.enabled", cmd.PersistentFlags().Lookup("chat.enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("chat.max_length", 512, "longest accepted chat message in characters, 0 for no limit")
	if err := viper.BindPFlag("chat.max_length", cmd.PersistentFlags().Lookup("chat.max_length")); err != nil {
		return err
	}

	return nil
}

func (s *Config) Set() {
	s.Enabled = viper.GetBool("chat.enabled")
	s.MaxLength = viper.GetInt("chat.max_length")
}
