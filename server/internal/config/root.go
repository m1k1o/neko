package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Root struct {
	Debug   bool
	Logs    bool
	CfgFile string
}

func (Root) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug mode")
	if err := viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug")); err != nil {
		return err
	}

	cmd.PersistentFlags().BoolP("logs", "l", false, "Save logs to file")
	if err := viper.BindPFlag("logs", cmd.PersistentFlags().Lookup("logs")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("config", "", "Configuration file path")
	if err := viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config")); err != nil {
		return err
	}

	return nil
}

func (s *Root) Set() {
	s.Logs = viper.GetBool("logs")
	s.Debug = viper.GetBool("debug")
	s.CfgFile = viper.GetString("config")
}
