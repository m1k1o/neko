package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Member struct {
	Provider      string
	FilePath      string
	Password      string
	AdminPassword string
}

func (Member) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("members_provider", "file", "choose members provider")
	if err := viper.BindPFlag("members_provider", cmd.PersistentFlags().Lookup("members_provider")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("members_file_path", "/home/neko/members.json", "mebmer file provider path")
	if err := viper.BindPFlag("members_file_path", cmd.PersistentFlags().Lookup("members_file_path")); err != nil {
		return err
	}

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

func (s *Member) Set() {
	s.Provider = viper.GetString("members_provider")
	s.FilePath = viper.GetString("members_file_path")
	s.Password = viper.GetString("password")
	s.AdminPassword = viper.GetString("password_admin")
}
