package config

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Session struct {
	File string

	PrivateMode       bool
	LockedLogins      bool
	LockedControls    bool
	ControlProtection bool
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

	cmd.PersistentFlags().Bool("session.private_mode", false, "whether private mode should be enabled initially")
	if err := viper.BindPFlag("session.private_mode", cmd.PersistentFlags().Lookup("session.private_mode")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("session.locked_logins", false, "whether logins should be locked for users initially")
	if err := viper.BindPFlag("session.locked_logins", cmd.PersistentFlags().Lookup("session.locked_logins")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("session.locked_controls", false, "whether controls should be locked for users initially")
	if err := viper.BindPFlag("session.locked_controls", cmd.PersistentFlags().Lookup("session.locked_controls")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("session.control_protection", false, "users can gain control only if at least one admin is in the room")
	if err := viper.BindPFlag("session.control_protection", cmd.PersistentFlags().Lookup("session.control_protection")); err != nil {
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

func (Session) InitV2(cmd *cobra.Command) error {
	cmd.PersistentFlags().StringSlice("locks", []string{}, "resources, that will be locked when starting (control, login)")
	if err := viper.BindPFlag("locks", cmd.PersistentFlags().Lookup("locks")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("control_protection", false, "control protection means, users can gain control only if at least one admin is in the room")
	if err := viper.BindPFlag("control_protection", cmd.PersistentFlags().Lookup("control_protection")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("implicit_control", false, "if enabled members can gain control implicitly")
	if err := viper.BindPFlag("implicit_control", cmd.PersistentFlags().Lookup("implicit_control")); err != nil {
		return err
	}

	return nil
}

func (s *Session) Set() {
	s.File = viper.GetString("session.file")

	s.PrivateMode = viper.GetBool("session.private_mode")
	s.LockedLogins = viper.GetBool("session.locked_logins")
	s.LockedControls = viper.GetBool("session.locked_controls")
	s.ControlProtection = viper.GetBool("session.control_protection")
	s.ImplicitHosting = viper.GetBool("session.implicit_hosting")
	s.InactiveCursors = viper.GetBool("session.inactive_cursors")
	s.MercifulReconnect = viper.GetBool("session.merciful_reconnect")
	s.APIToken = viper.GetString("session.api_token")

	s.CookieEnabled = viper.GetBool("session.cookie.enabled")
	s.CookieName = viper.GetString("session.cookie.name")
	s.CookieExpiration = time.Duration(viper.GetInt("session.cookie.expiration")) * time.Hour
	s.CookieSecure = viper.GetBool("session.cookie.secure")
}

func (s *Session) SetV2() {
	if viper.IsSet("locks") {
		locks := viper.GetStringSlice("locks")
		for _, lock := range locks {
			switch lock {
			// TODO: file_transfer
			case "control":
				s.LockedControls = true
			case "login":
				s.LockedLogins = true
			}
		}
		log.Warn().Msg("you are using v2 configuration 'NEKO_LOCKS' which is deprecated, please use 'NEKO_SESSION_LOCKED_CONTROLS' and 'NEKO_SESSION_LOCKED_LOGINS' instead")
	}

	if viper.IsSet("implicit_control") {
		s.ImplicitHosting = viper.GetBool("implicit_control")
		log.Warn().Msg("you are using v2 configuration 'NEKO_IMPLICIT_CONTROL' which is deprecated, please use 'NEKO_SESSION_IMPLICIT_HOSTING' instead")
	}
	if viper.IsSet("control_protection") {
		s.ControlProtection = viper.GetBool("control_protection")
		log.Warn().Msg("you are using v2 configuration 'NEKO_CONTROL_PROTECTION' which is deprecated, please use 'NEKO_SESSION_CONTROL_PROTECTION' instead")
	}
}
