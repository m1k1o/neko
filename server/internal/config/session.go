package config

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SessionCookie struct {
	Enabled    bool
	Name       string
	Expiration time.Duration
	Secure     bool
	HTTPOnly   bool
	Domain     string
	Path       string
}

type Session struct {
	File string

	PrivateMode       bool
	LockedLogins      bool
	LockedControls    bool
	ControlProtection bool
	ImplicitHosting   bool
	InactiveCursors   bool
	MercifulReconnect bool
	HeartbeatInterval int
	APIToken          string

	Cookie SessionCookie
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

	cmd.PersistentFlags().Int("session.heartbeat_interval", 10, "interval in seconds for sending heartbeat messages")
	if err := viper.BindPFlag("session.heartbeat_interval", cmd.PersistentFlags().Lookup("session.heartbeat_interval")); err != nil {
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

	cmd.PersistentFlags().Duration("session.cookie.expiration", 24*time.Hour, "expiration of the cookie")
	if err := viper.BindPFlag("session.cookie.expiration", cmd.PersistentFlags().Lookup("session.cookie.expiration")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("session.cookie.secure", true, "use secure cookies")
	if err := viper.BindPFlag("session.cookie.secure", cmd.PersistentFlags().Lookup("session.cookie.secure")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("session.cookie.http_only", true, "use http only cookies")
	if err := viper.BindPFlag("session.cookie.http_only", cmd.PersistentFlags().Lookup("session.cookie.http_only")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("session.cookie.domain", "", "domain of the cookie")
	if err := viper.BindPFlag("session.cookie.domain", cmd.PersistentFlags().Lookup("session.cookie.domain")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("session.cookie.path", "", "path of the cookie")
	if err := viper.BindPFlag("session.cookie.path", cmd.PersistentFlags().Lookup("session.cookie.path")); err != nil {
		return err
	}

	return nil
}

func (Session) InitV2(cmd *cobra.Command) error {
	cmd.PersistentFlags().StringSlice("locks", []string{}, "V2: resources, that will be locked when starting (control, login)")
	if err := viper.BindPFlag("locks", cmd.PersistentFlags().Lookup("locks")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("control_protection", false, "V2: control protection means, users can gain control only if at least one admin is in the room")
	if err := viper.BindPFlag("control_protection", cmd.PersistentFlags().Lookup("control_protection")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("implicit_control", false, "V2: if enabled members can gain control implicitly")
	if err := viper.BindPFlag("implicit_control", cmd.PersistentFlags().Lookup("implicit_control")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("heartbeat_interval", 120, "V2: heartbeat interval in seconds")
	if err := viper.BindPFlag("heartbeat_interval", cmd.PersistentFlags().Lookup("heartbeat_interval")); err != nil {
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
	s.HeartbeatInterval = viper.GetInt("session.heartbeat_interval")
	s.APIToken = viper.GetString("session.api_token")

	s.Cookie.Enabled = viper.GetBool("session.cookie.enabled")
	s.Cookie.Name = viper.GetString("session.cookie.name")
	s.Cookie.Expiration = viper.GetDuration("session.cookie.expiration")
	s.Cookie.Secure = viper.GetBool("session.cookie.secure")
	s.Cookie.HTTPOnly = viper.GetBool("session.cookie.http_only")
	s.Cookie.Domain = viper.GetString("session.cookie.domain")
	s.Cookie.Path = viper.GetString("session.cookie.path")
}

func (s *Session) SetV2() {
	enableLegacy := false

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
		enableLegacy = true
	}

	if viper.IsSet("implicit_control") {
		s.ImplicitHosting = viper.GetBool("implicit_control")
		log.Warn().Msg("you are using v2 configuration 'NEKO_IMPLICIT_CONTROL' which is deprecated, please use 'NEKO_SESSION_IMPLICIT_HOSTING' instead")
		enableLegacy = true
	}
	if viper.IsSet("control_protection") {
		s.ControlProtection = viper.GetBool("control_protection")
		log.Warn().Msg("you are using v2 configuration 'NEKO_CONTROL_PROTECTION' which is deprecated, please use 'NEKO_SESSION_CONTROL_PROTECTION' instead")
		enableLegacy = true
	}
	if viper.IsSet("heartbeat_interval") {
		s.HeartbeatInterval = viper.GetInt("heartbeat_interval")
		log.Warn().Msg("you are using v2 configuration 'NEKO_HEARTBEAT_INTERVAL' which is deprecated, please use 'NEKO_SESSION_HEARTBEAT_INTERVAL' instead")
		enableLegacy = true
	}

	// set legacy flag if any V2 configuration was used
	if !viper.IsSet("legacy") && enableLegacy {
		log.Warn().Msg("legacy configuration is enabled because at least one V2 configuration was used, please migrate to V3 configuration, visit https://neko.m1k1o.net/docs/v3/migration-from-v2 for more details")
		viper.Set("legacy", true)
	}
}
