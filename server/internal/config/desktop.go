package config

import (
	"os"
	"regexp"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/m1k1o/neko/server/pkg/types"
)

type Desktop struct {
	Display      string
	WindowID     uint64
	WindowX      int
	WindowY      int
	WindowWidth  int
	WindowHeight int

	ScreenSize types.ScreenSize

	UseInputDriver bool
	InputSocket    string

	Unminimize        bool
	UploadDrop        bool
	FileChooserDialog bool
}

func (Desktop) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("desktop.display", "", "X display to use for desktop sharing")
	if err := viper.BindPFlag("desktop.display", cmd.PersistentFlags().Lookup("desktop.display")); err != nil {
		return err
	}

	cmd.PersistentFlags().Uint64("desktop.window_id", 0, "X11 window ID to expose instead of the complete desktop")
	if err := viper.BindPFlag("desktop.window_id", cmd.PersistentFlags().Lookup("desktop.window_id")); err != nil {
		return err
	}
	cmd.PersistentFlags().Int("desktop.window_x", 0, "X coordinate of the desktop window region")
	if err := viper.BindPFlag("desktop.window_x", cmd.PersistentFlags().Lookup("desktop.window_x")); err != nil {
		return err
	}
	cmd.PersistentFlags().Int("desktop.window_y", 0, "Y coordinate of the desktop window region")
	if err := viper.BindPFlag("desktop.window_y", cmd.PersistentFlags().Lookup("desktop.window_y")); err != nil {
		return err
	}
	cmd.PersistentFlags().Int("desktop.window_width", 0, "width of the desktop window region")
	if err := viper.BindPFlag("desktop.window_width", cmd.PersistentFlags().Lookup("desktop.window_width")); err != nil {
		return err
	}
	cmd.PersistentFlags().Int("desktop.window_height", 0, "height of the desktop window region")
	if err := viper.BindPFlag("desktop.window_height", cmd.PersistentFlags().Lookup("desktop.window_height")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("desktop.screen", "1280x720@30", "default screen size and framerate")
	if err := viper.BindPFlag("desktop.screen", cmd.PersistentFlags().Lookup("desktop.screen")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("desktop.input.enabled", true, "whether custom xf86 input driver should be used to handle touchscreen")
	if err := viper.BindPFlag("desktop.input.enabled", cmd.PersistentFlags().Lookup("desktop.input.enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("desktop.input.socket", "/tmp/xf86-input-neko.sock", "socket path for custom xf86 input driver connection")
	if err := viper.BindPFlag("desktop.input.socket", cmd.PersistentFlags().Lookup("desktop.input.socket")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("desktop.unminimize", true, "automatically unminimize window when it is minimized")
	if err := viper.BindPFlag("desktop.unminimize", cmd.PersistentFlags().Lookup("desktop.unminimize")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("desktop.upload_drop", true, "whether drop upload is enabled")
	if err := viper.BindPFlag("desktop.upload_drop", cmd.PersistentFlags().Lookup("desktop.upload_drop")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("desktop.file_chooser_dialog", false, "whether to handle file chooser dialog externally")
	if err := viper.BindPFlag("desktop.file_chooser_dialog", cmd.PersistentFlags().Lookup("desktop.file_chooser_dialog")); err != nil {
		return err
	}

	return nil
}

func (Desktop) InitV2(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("screen", "", "V2: default screen resolution and framerate")
	if err := viper.BindPFlag("screen", cmd.PersistentFlags().Lookup("screen")); err != nil {
		return err
	}

	return nil
}

func (s *Desktop) Set() {
	s.Display = viper.GetString("desktop.display")

	// Display is provided by env variable unless explicitly set
	if s.Display == "" {
		s.Display = os.Getenv("DISPLAY")
	}
	s.WindowID = viper.GetUint64("desktop.window_id")
	s.WindowX = viper.GetInt("desktop.window_x")
	s.WindowY = viper.GetInt("desktop.window_y")
	s.WindowWidth = viper.GetInt("desktop.window_width")
	s.WindowHeight = viper.GetInt("desktop.window_height")

	s.ScreenSize = types.ScreenSize{
		Width:  1280,
		Height: 720,
		Rate:   30,
	}

	r := regexp.MustCompile(`([0-9]{1,4})x([0-9]{1,4})@([0-9]{1,3})`)
	res := r.FindStringSubmatch(viper.GetString("desktop.screen"))

	if len(res) > 0 {
		width, err1 := strconv.ParseInt(res[1], 10, 64)
		height, err2 := strconv.ParseInt(res[2], 10, 64)
		rate, err3 := strconv.ParseInt(res[3], 10, 64)

		if err1 == nil && err2 == nil && err3 == nil {
			s.ScreenSize.Width = int(width)
			s.ScreenSize.Height = int(height)
			s.ScreenSize.Rate = int16(rate)
		}
	}

	s.UseInputDriver = viper.GetBool("desktop.input.enabled")
	s.InputSocket = viper.GetString("desktop.input.socket")
	s.Unminimize = viper.GetBool("desktop.unminimize")
	s.UploadDrop = viper.GetBool("desktop.upload_drop")
	s.FileChooserDialog = viper.GetBool("desktop.file_chooser_dialog")
}

func (s *Desktop) SetV2() {
	enableLegacy := false

	if viper.IsSet("screen") {
		r := regexp.MustCompile(`([0-9]{1,4})x([0-9]{1,4})@([0-9]{1,3})`)
		res := r.FindStringSubmatch(viper.GetString("screen"))

		if len(res) > 0 {
			width, err1 := strconv.ParseInt(res[1], 10, 64)
			height, err2 := strconv.ParseInt(res[2], 10, 64)
			rate, err3 := strconv.ParseInt(res[3], 10, 64)

			if err1 == nil && err2 == nil && err3 == nil {
				s.ScreenSize.Width = int(width)
				s.ScreenSize.Height = int(height)
				s.ScreenSize.Rate = int16(rate)
			}
		}
		log.Warn().Msg("you are using v2 configuration 'NEKO_SCREEN' which is deprecated, please use 'NEKO_DESKTOP_SCREEN' instead")
		enableLegacy = true
	}

	// set legacy flag if any V2 configuration was used
	if !viper.IsSet("legacy") && enableLegacy {
		log.Warn().Msg("legacy configuration is enabled because at least one V2 configuration was used, please migrate to V3 configuration, visit https://neko.m1k1o.net/docs/v3/migration-from-v2 for more details")
		viper.Set("legacy", true)
	}
}
