package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Session struct {
	Password        string
	AdminPassword   string
	ImplicitHosting bool
	DatabaseAdapter string
	FilePath        string
}

func (Session) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("password", "neko", "password for connecting to stream")
	if err := viper.BindPFlag("password", cmd.PersistentFlags().Lookup("password")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("password_admin", "admin", "admin password for connecting to stream")
	if err := viper.BindPFlag("password_admin", cmd.PersistentFlags().Lookup("password_admin")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("implicit_hosting", true, "allow implicit control switching")
	if err := viper.BindPFlag("implicit_hosting", cmd.PersistentFlags().Lookup("implicit_hosting")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("database_adapter", "file", "choose database adapter for members")
	if err := viper.BindPFlag("database_adapter", cmd.PersistentFlags().Lookup("database_adapter")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("file_path", "/home/neko/members.json", "file adapter: specify file path")
	if err := viper.BindPFlag("file_path", cmd.PersistentFlags().Lookup("file_path")); err != nil {
		return err
	}

	return nil
}

func (s *Session) Set() {
	s.Password = viper.GetString("password")
	s.AdminPassword = viper.GetString("password_admin")
	s.ImplicitHosting = viper.GetBool("implicit_hosting")
	s.DatabaseAdapter = viper.GetString("database_adapter")
	s.FilePath = viper.GetString("file_path")
}
