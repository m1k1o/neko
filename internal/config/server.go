package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Server struct {
	Cert   string
	Key    string
	Bind   string
	Static string
	UserToken  string
	AdminToken string
}

func (Server) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("bind", "127.0.0.1:8080", "address/port/socket to serve neko")
	if err := viper.BindPFlag("bind", cmd.PersistentFlags().Lookup("bind")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("cert", "", "path to the SSL cert used to secure the neko server")
	if err := viper.BindPFlag("cert", cmd.PersistentFlags().Lookup("cert")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("key", "", "path to the SSL key used to secure the neko server")
	if err := viper.BindPFlag("key", cmd.PersistentFlags().Lookup("key")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("static", "", "path to neko client files to serve")
	if err := viper.BindPFlag("static", cmd.PersistentFlags().Lookup("static")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("user_token", "user_secret", "JWT token for users")
	if err := viper.BindPFlag("user_token", cmd.PersistentFlags().Lookup("user_token")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("admin_token", "admin_secret", "JWT token for admins")
	if err := viper.BindPFlag("admin_token", cmd.PersistentFlags().Lookup("admin_token")); err != nil {
		return err
	}

	return nil
}

func (s *Server) Set() {
	s.Cert = viper.GetString("cert")
	s.Key = viper.GetString("key")
	s.Bind = viper.GetString("bind")
	s.Static = viper.GetString("static")
	s.UserToken = viper.GetString("user_token")
	s.AdminToken = viper.GetString("admin_token")
}
