package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Session struct {
	ImplicitHosting bool
}

func (Session) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("implicit_hosting", true, "allow implicit control switching")
	if err := viper.BindPFlag("implicit_hosting", cmd.PersistentFlags().Lookup("implicit_hosting")); err != nil {
		return err
	}

	return nil
}

func (s *Session) Set() {
	s.ImplicitHosting = viper.GetBool("implicit_hosting")
}
