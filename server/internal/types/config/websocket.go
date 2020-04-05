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
	cmd.PersistentFlags().String("password", "neko", "password for connecting to stream")
	if err := viper.BindPFlag("password", cmd.PersistentFlags().Lookup("password")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("password_admin", "admin", "admin password for connecting to stream")
	if err := viper.BindPFlag("password_admin", cmd.PersistentFlags().Lookup("password_admin")); err != nil {
		return err
	}

	return nil
}

func (s *WebSocket) Set() {
	s.Password = viper.GetString("password")
	s.AdminPassword = viper.GetString("password_admin")
}
