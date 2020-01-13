package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Serve struct {
	Cert     string
	Key      string
	Bind     string
	Password string
	Static string
}

func (Serve) Init(cmd *cobra.Command) error {

	cmd.PersistentFlags().String("bind", "127.0.0.1:8080", "Address/port/socket to serve neko")
	if err := viper.BindPFlag("bind", cmd.PersistentFlags().Lookup("bind")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("cert", "", "Path to the SSL cert used to secure the neko server")
	if err := viper.BindPFlag("cert", cmd.PersistentFlags().Lookup("cert")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("key", "", "Path to the SSL key used to secure the neko server")
	if err := viper.BindPFlag("key", cmd.PersistentFlags().Lookup("key")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("password", "neko", "Password for connecting to stream")
	if err := viper.BindPFlag("password", cmd.PersistentFlags().Lookup("password")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("static", "./www", "Static files to serve")
	if err := viper.BindPFlag("static", cmd.PersistentFlags().Lookup("static")); err != nil {
		return err
	}

	return nil
}

func (s *Serve) Set() {
	s.Cert = viper.GetString("cert")
	s.Key = viper.GetString("key")
	s.Bind = viper.GetString("bind")
	s.Password = viper.GetString("password")
	s.Static = viper.GetString("static")
}
