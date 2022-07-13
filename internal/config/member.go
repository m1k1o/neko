package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/demodesk/neko/internal/member/file"
	"github.com/demodesk/neko/internal/member/object"
)

type Member struct {
	Provider string

	// providers
	File   file.Config
	Object object.Config
}

func (Member) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("member.provider", "object", "choose member provider")
	if err := viper.BindPFlag("member.provider", cmd.PersistentFlags().Lookup("member.provider")); err != nil {
		return err
	}

	// file provider
	cmd.PersistentFlags().String("member.file.path", "", "member file provider: storage path")
	if err := viper.BindPFlag("member.file.path", cmd.PersistentFlags().Lookup("member.file.path")); err != nil {
		return err
	}

	// object provider
	cmd.PersistentFlags().String("member.object.user_password", "", "member object provider: user password")
	if err := viper.BindPFlag("member.object.user_password", cmd.PersistentFlags().Lookup("member.object.user_password")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.object.admin_password", "", "member object provider: admin password")
	if err := viper.BindPFlag("member.object.admin_password", cmd.PersistentFlags().Lookup("member.object.admin_password")); err != nil {
		return err
	}

	return nil
}

func (s *Member) Set() {
	s.Provider = viper.GetString("member.provider")

	// file provider
	s.File.Path = viper.GetString("member.file.path")

	// object provider
	s.Object.UserPassword = viper.GetString("member.object.user_password")
	s.Object.AdminPassword = viper.GetString("member.object.admin_password")
}
