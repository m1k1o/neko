package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Session struct {
	ImplicitHosting bool
	APIToken        string
}

func (Session) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("session.implicit_hosting", true, "allow implicit control switching")
	if err := viper.BindPFlag("session.implicit_hosting", cmd.PersistentFlags().Lookup("session.implicit_hosting")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("session.api_token", "", "API token for interacting with external services")
	if err := viper.BindPFlag("session.api_token", cmd.PersistentFlags().Lookup("session.api_token")); err != nil {
		return err
	}

	return nil
}

func (s *Session) Set() {
	s.ImplicitHosting = viper.GetBool("session.implicit_hosting")
	s.APIToken = viper.GetString("session.api_token")
}
