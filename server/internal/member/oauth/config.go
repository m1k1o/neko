package oauth

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

// Config contains the OAuth member provider settings.
type Config struct {
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
	AdminProfile     types.MemberProfile
	UserProfile      types.MemberProfile
}

func Init(cmd *cobra.Command) error {
	flags := cmd.PersistentFlags()
	flags.Bool("member.oauth.enabled", false, "enable OAuth 2.0 login")
	if err := viper.BindPFlag("member.oauth.enabled", flags.Lookup("member.oauth.enabled")); err != nil {
		return err
	}
	flags.Bool("member.oauth.auto_redirect", false, "redirect the Neko root page to OAuth login automatically")
	if err := viper.BindPFlag("member.oauth.auto_redirect", flags.Lookup("member.oauth.auto_redirect")); err != nil {
		return err
	}
	flags.String("member.oauth.name", "OAuth", "display name for the OAuth login option")
	if err := viper.BindPFlag("member.oauth.name", flags.Lookup("member.oauth.name")); err != nil {
		return err
	}
	flags.StringSlice("member.oauth.admin_emails", []string{}, "OAuth user-info email addresses granted administrator access")
	if err := viper.BindPFlag("member.oauth.admin_emails", flags.Lookup("member.oauth.admin_emails")); err != nil {
		return err
	}
	flags.String("member.oauth.client_id", "", "OAuth 2.0 client ID")
	if err := viper.BindPFlag("member.oauth.client_id", flags.Lookup("member.oauth.client_id")); err != nil {
		return err
	}
	flags.String("member.oauth.client_secret", "", "OAuth 2.0 client secret")
	if err := viper.BindPFlag("member.oauth.client_secret", flags.Lookup("member.oauth.client_secret")); err != nil {
		return err
	}
	flags.String("member.oauth.issuer_url", "", "OpenID Connect issuer URL used to discover OAuth endpoints")
	if err := viper.BindPFlag("member.oauth.issuer_url", flags.Lookup("member.oauth.issuer_url")); err != nil {
		return err
	}
	flags.String("member.oauth.authorization_url", "", "OAuth 2.0 authorization endpoint")
	if err := viper.BindPFlag("member.oauth.authorization_url", flags.Lookup("member.oauth.authorization_url")); err != nil {
		return err
	}
	flags.String("member.oauth.token_url", "", "OAuth 2.0 token endpoint")
	if err := viper.BindPFlag("member.oauth.token_url", flags.Lookup("member.oauth.token_url")); err != nil {
		return err
	}
	flags.String("member.oauth.userinfo_url", "", "OAuth 2.0 user-info endpoint")
	if err := viper.BindPFlag("member.oauth.userinfo_url", flags.Lookup("member.oauth.userinfo_url")); err != nil {
		return err
	}
	flags.String("member.oauth.redirect_url", "", "OAuth 2.0 callback URL registered with the provider")
	if err := viper.BindPFlag("member.oauth.redirect_url", flags.Lookup("member.oauth.redirect_url")); err != nil {
		return err
	}
	flags.StringSlice("member.oauth.scopes", []string{"openid", "profile"}, "OAuth 2.0 scopes")
	if err := viper.BindPFlag("member.oauth.scopes", flags.Lookup("member.oauth.scopes")); err != nil {
		return err
	}
	flags.String("member.oauth.subject_field", "sub", "field in user-info response used as the stable user identifier")
	if err := viper.BindPFlag("member.oauth.subject_field", flags.Lookup("member.oauth.subject_field")); err != nil {
		return err
	}
	flags.String("member.oauth.username_field", "name", "field in user-info response used as the display name")
	if err := viper.BindPFlag("member.oauth.username_field", flags.Lookup("member.oauth.username_field")); err != nil {
		return err
	}
	flags.String("member.oauth.avatar_field", "picture", "field in user-info response used as the avatar URL")
	if err := viper.BindPFlag("member.oauth.avatar_field", flags.Lookup("member.oauth.avatar_field")); err != nil {
		return err
	}
	flags.String("member.oauth.success_redirect", "/", "relative URL to redirect to after OAuth login")
	if err := viper.BindPFlag("member.oauth.success_redirect", flags.Lookup("member.oauth.success_redirect")); err != nil {
		return err
	}
	flags.String("member.oauth.user_profile", "{}", "OAuth user permission profile")
	if err := viper.BindPFlag("member.oauth.user_profile", flags.Lookup("member.oauth.user_profile")); err != nil {
		return err
	}
	flags.String("member.oauth.admin_profile", "{}", "OAuth administrator permission profile")
	return viper.BindPFlag("member.oauth.admin_profile", flags.Lookup("member.oauth.admin_profile"))
}

func (c *Config) Set(userProfile, adminProfile types.MemberProfile) error {
	c.Enabled = viper.GetBool("member.oauth.enabled")
	c.AutoRedirect = viper.GetBool("member.oauth.auto_redirect")
	c.Name = viper.GetString("member.oauth.name")
	c.AdminEmails = viper.GetStringSlice("member.oauth.admin_emails")
	c.ClientID = viper.GetString("member.oauth.client_id")
	c.ClientSecret = viper.GetString("member.oauth.client_secret")
	c.IssuerURL = viper.GetString("member.oauth.issuer_url")
	c.AuthorizationURL = viper.GetString("member.oauth.authorization_url")
	c.TokenURL = viper.GetString("member.oauth.token_url")
	c.UserInfoURL = viper.GetString("member.oauth.userinfo_url")
	c.RedirectURL = viper.GetString("member.oauth.redirect_url")
	c.Scopes = viper.GetStringSlice("member.oauth.scopes")
	c.SubjectField = viper.GetString("member.oauth.subject_field")
	c.UsernameField = viper.GetString("member.oauth.username_field")
	c.AvatarField = viper.GetString("member.oauth.avatar_field")
	c.SuccessRedirect = viper.GetString("member.oauth.success_redirect")
	c.UserProfile = userProfile
	c.AdminProfile = adminProfile
	if err := viper.UnmarshalKey("member.oauth.user_profile", &c.UserProfile, viper.DecodeHook(utils.JsonStringAutoDecode(c.UserProfile))); err != nil {
		return err
	}
	return viper.UnmarshalKey("member.oauth.admin_profile", &c.AdminProfile, viper.DecodeHook(utils.JsonStringAutoDecode(c.AdminProfile)))
}
