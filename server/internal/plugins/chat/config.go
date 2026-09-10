package chat

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Enabled      bool
	HistoryFile  string
	HistoryLimit int
}

func (Config) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("chat.enabled", true, "whether to enable chat plugin")
	if err := viper.BindPFlag("chat.enabled", cmd.PersistentFlags().Lookup("chat.enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("chat.history_file", "/home/neko/.local/share/neko/chat-history.json", "path to the persistent chat history file")
	if err := viper.BindPFlag("chat.history_file", cmd.PersistentFlags().Lookup("chat.history_file")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("chat.history_limit", 200, "maximum number of text messages to retain")
	if err := viper.BindPFlag("chat.history_limit", cmd.PersistentFlags().Lookup("chat.history_limit")); err != nil {
		return err
	}

	return nil
}

func (s *Config) Set() {
	s.Enabled = viper.GetBool("chat.enabled")
	s.HistoryFile = viper.GetString("chat.history_file")
	if s.HistoryFile != "" {
		s.HistoryFile = filepath.Clean(s.HistoryFile)
	}
	s.HistoryLimit = viper.GetInt("chat.history_limit")
	if s.HistoryLimit < 0 {
		s.HistoryLimit = 0
	}
}
