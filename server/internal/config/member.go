package config

import (
	"strings"

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
	OAuth     OAuth
}

// OAuth contains the generic OAuth 2.0 authorization-code flow settings.
// The provider's user-info response is mapped to a Neko member profile.
type OAuth struct {
	Enabled          bool
	AutoRedirect     bool
	Name             string
	AdminEmails      []string
	ClientID         string
	ClientSecret     string
	IssuerURL        string
	AuthorizationURL string
	TokenURL         string
	UserInfoURL      string
	RedirectURL      string
	Scopes           []string
	SubjectField     string
	UsernameField    string
	AvatarField      string
	SuccessRedirect  string
	UserProfile      types.MemberProfile
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

	// OAuth 2.0 provider
	cmd.PersistentFlags().Bool("member.oauth.enabled", false, "enable OAuth 2.0 login")
	if err := viper.BindPFlag("member.oauth.enabled", cmd.PersistentFlags().Lookup("member.oauth.enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("member.oauth.auto_redirect", false, "redirect the Neko root page to OAuth login automatically")
	if err := viper.BindPFlag("member.oauth.auto_redirect", cmd.PersistentFlags().Lookup("member.oauth.auto_redirect")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.name", "OAuth", "display name for the OAuth login option")
	if err := viper.BindPFlag("member.oauth.name", cmd.PersistentFlags().Lookup("member.oauth.name")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.admin_email", "", "comma-separated OAuth user-info email addresses granted administrator access")
	if err := viper.BindPFlag("member.oauth.admin_email", cmd.PersistentFlags().Lookup("member.oauth.admin_email")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.client_id", "", "OAuth 2.0 client ID")
	if err := viper.BindPFlag("member.oauth.client_id", cmd.PersistentFlags().Lookup("member.oauth.client_id")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.client_secret", "", "OAuth 2.0 client secret")
	if err := viper.BindPFlag("member.oauth.client_secret", cmd.PersistentFlags().Lookup("member.oauth.client_secret")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.issuer_url", "", "OpenID Connect issuer URL used to discover OAuth endpoints")
	if err := viper.BindPFlag("member.oauth.issuer_url", cmd.PersistentFlags().Lookup("member.oauth.issuer_url")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.authorization_url", "", "OAuth 2.0 authorization endpoint")
	if err := viper.BindPFlag("member.oauth.authorization_url", cmd.PersistentFlags().Lookup("member.oauth.authorization_url")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.token_url", "", "OAuth 2.0 token endpoint")
	if err := viper.BindPFlag("member.oauth.token_url", cmd.PersistentFlags().Lookup("member.oauth.token_url")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.userinfo_url", "", "OAuth 2.0 user-info endpoint")
	if err := viper.BindPFlag("member.oauth.userinfo_url", cmd.PersistentFlags().Lookup("member.oauth.userinfo_url")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.redirect_url", "", "OAuth 2.0 callback URL registered with the provider")
	if err := viper.BindPFlag("member.oauth.redirect_url", cmd.PersistentFlags().Lookup("member.oauth.redirect_url")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("member.oauth.scopes", []string{"openid", "profile"}, "OAuth 2.0 scopes")
	if err := viper.BindPFlag("member.oauth.scopes", cmd.PersistentFlags().Lookup("member.oauth.scopes")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.subject_field", "sub", "field in user-info response used as the stable user identifier")
	if err := viper.BindPFlag("member.oauth.subject_field", cmd.PersistentFlags().Lookup("member.oauth.subject_field")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.username_field", "name", "field in user-info response used as the display name")
	if err := viper.BindPFlag("member.oauth.username_field", cmd.PersistentFlags().Lookup("member.oauth.username_field")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.avatar_field", "picture", "field in user-info response used as the avatar URL")
	if err := viper.BindPFlag("member.oauth.avatar_field", cmd.PersistentFlags().Lookup("member.oauth.avatar_field")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.success_redirect", "/", "relative URL to redirect to after OAuth login")
	if err := viper.BindPFlag("member.oauth.success_redirect", cmd.PersistentFlags().Lookup("member.oauth.success_redirect")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("member.oauth.user_profile", "{}", "OAuth user permission profile")
	if err := viper.BindPFlag("member.oauth.user_profile", cmd.PersistentFlags().Lookup("member.oauth.user_profile")); err != nil {
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

	// OAuth provider
	s.OAuth.Enabled = viper.GetBool("member.oauth.enabled")
	s.OAuth.AutoRedirect = viper.GetBool("member.oauth.auto_redirect")
	s.OAuth.Name = viper.GetString("member.oauth.name")
	s.OAuth.AdminEmails = commaSeparatedValues(viper.GetString("member.oauth.admin_email"))
	s.OAuth.ClientID = viper.GetString("member.oauth.client_id")
	s.OAuth.ClientSecret = viper.GetString("member.oauth.client_secret")
	s.OAuth.IssuerURL = viper.GetString("member.oauth.issuer_url")
	s.OAuth.AuthorizationURL = viper.GetString("member.oauth.authorization_url")
	s.OAuth.TokenURL = viper.GetString("member.oauth.token_url")
	s.OAuth.UserInfoURL = viper.GetString("member.oauth.userinfo_url")
	s.OAuth.RedirectURL = viper.GetString("member.oauth.redirect_url")
	s.OAuth.Scopes = viper.GetStringSlice("member.oauth.scopes")
	s.OAuth.SubjectField = viper.GetString("member.oauth.subject_field")
	s.OAuth.UsernameField = viper.GetString("member.oauth.username_field")
	s.OAuth.AvatarField = viper.GetString("member.oauth.avatar_field")
	s.OAuth.SuccessRedirect = viper.GetString("member.oauth.success_redirect")

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
	s.OAuth.UserProfile = s.Multiuser.UserProfile

	// override user profile
	if err := viper.UnmarshalKey("member.multiuser.user_profile", &s.Multiuser.UserProfile, viper.DecodeHook(
		utils.JsonStringAutoDecode(s.Multiuser.UserProfile),
	)); err != nil {
		log.Warn().Err(err).Msgf("unable to parse member multiuser user profile")
	}
	if err := viper.UnmarshalKey("member.oauth.user_profile", &s.OAuth.UserProfile, viper.DecodeHook(
		utils.JsonStringAutoDecode(s.OAuth.UserProfile),
	)); err != nil {
		log.Warn().Err(err).Msgf("unable to parse member OAuth user profile")
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

func commaSeparatedValues(value string) []string {
	values := strings.Split(value, ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
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
