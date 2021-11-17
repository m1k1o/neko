package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type WebSocket struct {
	Password      string
	AdminPassword string
	Proxy         bool
	Locks         []string

	ControlProtection bool
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

	cmd.PersistentFlags().Bool("proxy", false, "enable reverse proxy mode")
	if err := viper.BindPFlag("proxy", cmd.PersistentFlags().Lookup("proxy")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("locks", []string{}, "resources, that will be locked when starting (control, login)")
	if err := viper.BindPFlag("locks", cmd.PersistentFlags().Lookup("locks")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("control_protection", false, "control protection means, users can gain control only if at least one admin is in the room")
	if err := viper.BindPFlag("control_protection", cmd.PersistentFlags().Lookup("control_protection")); err != nil {
		return err
	}

	return nil
}

func (s *WebSocket) Set() {
	s.Password = viper.GetString("password")
	s.AdminPassword = viper.GetString("password_admin")
	s.Proxy = viper.GetBool("proxy")
	s.Locks = viper.GetStringSlice("locks")

	s.ControlProtection = viper.GetBool("control_protection")
}
