package config

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Session struct {
	File string

	ImplicitHosting   bool
	InactiveCursors   bool
	MercifulReconnect bool
	APIToken          string

	CookieEnabled    bool
	CookieName       string
	CookieExpiration time.Duration
	CookieSecure     bool
}

func (Session) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("session.file", "", "if sessions should be stored in a file, otherwise they will be stored only in memory")
	if err := viper.BindPFlag("session.file", cmd.PersistentFlags().Lookup("session.file")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("session.implicit_hosting", true, "allow implicit control switching")
	if err := viper.BindPFlag("session.implicit_hosting", cmd.PersistentFlags().Lookup("session.implicit_hosting")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("session.inactive_cursors", false, "show inactive cursors on the screen")
	if err := viper.BindPFlag("session.inactive_cursors", cmd.PersistentFlags().Lookup("session.inactive_cursors")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("session.merciful_reconnect", true, "allow reconnecting to websocket even if previous connection was not closed")
	if err := viper.BindPFlag("session.merciful_reconnect", cmd.PersistentFlags().Lookup("session.merciful_reconnect")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("session.api_token", "", "API token for interacting with external services")
	if err := viper.BindPFlag("session.api_token", cmd.PersistentFlags().Lookup("session.api_token")); err != nil {
		return err
	}

	// cookie
	cmd.PersistentFlags().Bool("session.cookie.enabled", true, "whether cookies authentication should be enabled")
	if err := viper.BindPFlag("session.cookie.enabled", cmd.PersistentFlags().Lookup("session.cookie.enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("session.cookie.name", "NEKO_SESSION", "name of the cookie that holds token")
	if err := viper.BindPFlag("session.cookie.name", cmd.PersistentFlags().Lookup("session.cookie.name")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("session.cookie.expiration", 365*24, "expiration of the cookie in hours")
	if err := viper.BindPFlag("session.cookie.expiration", cmd.PersistentFlags().Lookup("session.cookie.expiration")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("session.cookie.secure", true, "use secure cookies")
	if err := viper.BindPFlag("session.cookie.secure", cmd.PersistentFlags().Lookup("session.cookie.secure")); err != nil {
		return err
	}

	return nil
}

func (s *Session) Set() {
	s.File = viper.GetString("session.file")

	s.ImplicitHosting = viper.GetBool("session.implicit_hosting")
	s.InactiveCursors = viper.GetBool("session.inactive_cursors")
	s.MercifulReconnect = viper.GetBool("session.merciful_reconnect")
	s.APIToken = viper.GetString("session.api_token")

	s.CookieEnabled = viper.GetBool("session.cookie.enabled")
	s.CookieName = viper.GetString("session.cookie.name")
	s.CookieExpiration = time.Duration(viper.GetInt("session.cookie.expiration")) * time.Hour
	s.CookieSecure = viper.GetBool("session.cookie.secure")
}
