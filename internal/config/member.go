package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/demodesk/neko/internal/member/file"
	"github.com/demodesk/neko/internal/member/multiuser"
	"github.com/demodesk/neko/internal/member/object"
	"github.com/demodesk/neko/pkg/utils"
)

type Member struct {
	Provider string

	// providers
	File      file.Config
	Object    object.Config
	Multiuser multiuser.Config
}

func (Member) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("member.provider", "multiuser", "choose member provider")
	if err := viper.BindPFlag("member.provider", cmd.PersistentFlags().Lookup("member.provider")); err != nil {
		return err
	}

	// file provider
	cmd.PersistentFlags().String("member.file.path", "", "member file provider: storage path")
	if err := viper.BindPFlag("member.file.path", cmd.PersistentFlags().Lookup("member.file.path")); err != nil {
		return err
	}

	// object provider
	cmd.PersistentFlags().String("member.object.users", "[]", "member object provider: users in JSON format")
	if err := viper.BindPFlag("member.object.users", cmd.PersistentFlags().Lookup("member.object.users")); err != nil {
		return err
	}

	// multiuser provider
	cmd.PersistentFlags().String("member.multiuser.user_password", "neko", "member multiuser provider: user password")
	if err := viper.BindPFlag("member.multiuser.user_password", cmd.PersistentFlags().Lookup("member.multiuser.user_password")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.multiuser.admin_password", "admin", "member multiuser provider: admin password")
	if err := viper.BindPFlag("member.multiuser.admin_password", cmd.PersistentFlags().Lookup("member.multiuser.admin_password")); err != nil {
		return err
	}

	return nil
}

func (s *Member) Set() {
	s.Provider = viper.GetString("member.provider")

	// file provider
	s.File.Path = viper.GetString("member.file.path")

	// object provider
	if err := viper.UnmarshalKey("member.object.users", &s.Object.Users, viper.DecodeHook(
		utils.JsonStringAutoDecode(s.Object.Users),
	)); err != nil {
		log.Warn().Err(err).Msgf("unable to parse member object users")
	}

	// multiuser provider
	s.Multiuser.UserPassword = viper.GetString("member.multiuser.user_password")
	s.Multiuser.AdminPassword = viper.GetString("member.multiuser.admin_password")
}
