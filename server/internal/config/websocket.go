package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type WebSocket struct {
	Password      string
	AdminPassword string
}

func (WebSocket) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("password", "neko", "Password for connecting to stream")
	if err := viper.BindPFlag("password", cmd.PersistentFlags().Lookup("password")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("admin", "admin", "Admin password for connecting to stream")
	if err := viper.BindPFlag("admin", cmd.PersistentFlags().Lookup("admin")); err != nil {
		return err
	}

	return nil
}

func (s *WebSocket) Set() {
	s.Password = viper.GetString("password")
	s.AdminPassword = viper.GetString("admin")
}
