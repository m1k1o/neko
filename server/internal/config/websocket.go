package config

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type WebSocket struct {
	Password      string
	AdminPassword string
	Locks         []string

	ControlProtection bool

	FileTransferEnabled bool
	FileTransferPath    string
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

	cmd.PersistentFlags().StringSlice("locks", []string{}, "resources, that will be locked when starting (control, login)")
	if err := viper.BindPFlag("locks", cmd.PersistentFlags().Lookup("locks")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("control_protection", false, "control protection means, users can gain control only if at least one admin is in the room")
	if err := viper.BindPFlag("control_protection", cmd.PersistentFlags().Lookup("control_protection")); err != nil {
		return err
	}

	// File transfer

	cmd.PersistentFlags().Bool("file_transfer_enabled", false, "enable file transfer feature")
	if err := viper.BindPFlag("file_transfer_enabled", cmd.PersistentFlags().Lookup("file_transfer_enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("file_transfer_path", "/home/neko/Downloads", "path to use for file transfer")
	if err := viper.BindPFlag("file_transfer_path", cmd.PersistentFlags().Lookup("file_transfer_path")); err != nil {
		return err
	}

	return nil
}

func (s *WebSocket) Set() {
	s.Password = viper.GetString("password")
	s.AdminPassword = viper.GetString("password_admin")
	s.Locks = viper.GetStringSlice("locks")

	s.ControlProtection = viper.GetBool("control_protection")

	s.FileTransferEnabled = viper.GetBool("file_transfer_enabled")
	s.FileTransferPath = viper.GetString("file_transfer_path")
	s.FileTransferPath = filepath.Clean(s.FileTransferPath)
}
