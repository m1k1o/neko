package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/m1k1o/neko/server/internal/member/file"
	"github.com/m1k1o/neko/server/internal/member/multiuser"
	"github.com/m1k1o/neko/server/internal/member/object"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

type Member struct {
	Provider string

	// providers
	File      file.Config
	Object    object.Config
	Multiuser multiuser.Config
}

func (Member) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("member.provider", "multiuser", "selected member provider")
	if err := viper.BindPFlag("member.provider", cmd.PersistentFlags().Lookup("member.provider")); err != nil {
		return err
	}

	// file provider
	cmd.PersistentFlags().String("member.file.path", "", "member file provider: path to the file containing the users and their passwords")
	if err := viper.BindPFlag("member.file.path", cmd.PersistentFlags().Lookup("member.file.path")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("member.file.hash", true, "member file provider: whether the passwords are hashed using sha256 or not (recommended)")
	if err := viper.BindPFlag("member.file.hash", cmd.PersistentFlags().Lookup("member.file.hash")); err != nil {
		return err
	}

	// object provider
	cmd.PersistentFlags().String("member.object.users", "[]", "member object provider: list of users with their passwords and profiles")
	if err := viper.BindPFlag("member.object.users", cmd.PersistentFlags().Lookup("member.object.users")); err != nil {
		return err
	}

	// multiuser provider
	cmd.PersistentFlags().String("member.multiuser.user_password", "neko", "member multiuser provider: password for regular users")
	if err := viper.BindPFlag("member.multiuser.user_password", cmd.PersistentFlags().Lookup("member.multiuser.user_password")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.multiuser.admin_password", "admin", "member multiuser provider: password for admin users")
	if err := viper.BindPFlag("member.multiuser.admin_password", cmd.PersistentFlags().Lookup("member.multiuser.admin_password")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.multiuser.user_profile", "{}", "member multiuser provider: profile template for regular users")
	if err := viper.BindPFlag("member.multiuser.user_profile", cmd.PersistentFlags().Lookup("member.multiuser.user_profile")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.multiuser.admin_profile", "{}", "member multiuser provider: profile template for admin users")
	if err := viper.BindPFlag("member.multiuser.admin_profile", cmd.PersistentFlags().Lookup("member.multiuser.admin_profile")); err != nil {
		return err
	}

	return nil
}

func (Member) InitV2(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("password", "", "V2: password for connecting to stream")
	if err := viper.BindPFlag("password", cmd.PersistentFlags().Lookup("password")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("password_admin", "", "V2: admin password for connecting to stream")
	if err := viper.BindPFlag("password_admin", cmd.PersistentFlags().Lookup("password_admin")); err != nil {
		return err
	}

	return nil
}

func (s *Member) Set() {
	s.Provider = viper.GetString("member.provider")

	// file provider
	s.File.Path = viper.GetString("member.file.path")
	s.File.Hash = viper.GetBool("member.file.hash")

	// object provider
	if err := viper.UnmarshalKey("member.object.users", &s.Object.Users, viper.DecodeHook(
		utils.JsonStringAutoDecode(s.Object.Users),
	)); err != nil {
		log.Warn().Err(err).Msgf("unable to parse member object users")
	}

	// multiuser provider
	s.Multiuser.UserPassword = viper.GetString("member.multiuser.user_password")
	s.Multiuser.AdminPassword = viper.GetString("member.multiuser.admin_password")

	// default user profile
	s.Multiuser.UserProfile = types.MemberProfile{
		IsAdmin:               false,
		CanLogin:              true,
		CanConnect:            true,
		CanWatch:              true,
		CanHost:               true,
		CanShareMedia:         true,
		CanAccessClipboard:    true,
		SendsInactiveCursor:   true,
		CanSeeInactiveCursors: false,
	}

	// override user profile
	if err := viper.UnmarshalKey("member.multiuser.user_profile", &s.Multiuser.UserProfile, viper.DecodeHook(
		utils.JsonStringAutoDecode(s.Multiuser.UserProfile),
	)); err != nil {
		log.Warn().Err(err).Msgf("unable to parse member multiuser user profile")
	}

	// default admin profile
	s.Multiuser.AdminProfile = types.MemberProfile{
		IsAdmin:               true,
		CanLogin:              true,
		CanConnect:            true,
		CanWatch:              true,
		CanHost:               true,
		CanShareMedia:         true,
		CanAccessClipboard:    true,
		SendsInactiveCursor:   true,
		CanSeeInactiveCursors: true,
	}

	// override admin profile
	if err := viper.UnmarshalKey("member.multiuser.admin_profile", &s.Multiuser.AdminProfile, viper.DecodeHook(
		utils.JsonStringAutoDecode(s.Multiuser.AdminProfile),
	)); err != nil {
		log.Warn().Err(err).Msgf("unable to parse member multiuser admin profile")
	}
}

func (s *Member) SetV2() {
	enableLegacy := false

	if viper.IsSet("password") || viper.IsSet("password_admin") {
		s.Provider = "multiuser"
		if userPassword := viper.GetString("password"); userPassword != "" {
			s.Multiuser.UserPassword = userPassword
		} else {
			s.Multiuser.UserPassword = "neko"
		}
		if adminPassword := viper.GetString("password_admin"); adminPassword != "" {
			s.Multiuser.AdminPassword = adminPassword
		} else {
			s.Multiuser.AdminPassword = "admin"
		}
		log.Warn().Msg("you are using v2 configuration 'NEKO_PASSWORD' and 'NEKO_PASSWORD_ADMIN' which are deprecated, please use 'NEKO_MEMBER_MULTIUSER_USER_PASSWORD' and 'NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD' with 'NEKO_MEMBER_PROVIDER=multiuser' instead")
		enableLegacy = true
	}

	// set legacy flag if any V2 configuration was used
	if !viper.IsSet("legacy") && enableLegacy {
		log.Warn().Msg("legacy configuration is enabled because at least one V2 configuration was used, please migrate to V3 configuration, visit https://neko.m1k1o.net/docs/v3/migration-from-v2 for more details")
		viper.Set("legacy", true)
	}
}
