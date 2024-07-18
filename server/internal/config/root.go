package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Root struct {
	Config string

	LogLevel   zerolog.Level
	LogTime    string
	LogJson    bool
	LogNocolor bool
	LogDir     string
}

func (Root) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().StringP("config", "c", "", "configuration file path")
	if err := viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config")); err != nil {
		return err
	}

	// just a shortcut
	cmd.PersistentFlags().BoolP("debug", "d", false, "enable debug mode")
	if err := viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("log.level", "info", "set log level (trace, debug, info, warn, error, fatal, panic, disabled)")
	if err := viper.BindPFlag("log.level", cmd.PersistentFlags().Lookup("log.level")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("log.time", "unix", "time format used in logs (unix, unixms, unixmicro)")
	if err := viper.BindPFlag("log.time", cmd.PersistentFlags().Lookup("log.time")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("log.json", false, "logs in JSON format")
	if err := viper.BindPFlag("log.json", cmd.PersistentFlags().Lookup("log.json")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("log.nocolor", false, "no ANSI colors in non-JSON output")
	if err := viper.BindPFlag("log.nocolor", cmd.PersistentFlags().Lookup("log.nocolor")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("log.dir", "", "logging directory to store logs")
	if err := viper.BindPFlag("log.dir", cmd.PersistentFlags().Lookup("log.dir")); err != nil {
		return err
	}

	return nil
}

func (Root) InitV2(cmd *cobra.Command) error {
	cmd.PersistentFlags().BoolP("logs", "l", false, "save logs to file")
	if err := viper.BindPFlag("logs", cmd.PersistentFlags().Lookup("logs")); err != nil {
		return err
	}

	return nil
}

func (s *Root) Set() {
	s.Config = viper.GetString("config")

	logLevel := viper.GetString("log.level")
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Warn().Msgf("unknown log level %s", logLevel)
	} else {
		s.LogLevel = level
	}

	logTime := viper.GetString("log.time")
	switch logTime {
	case "unix":
		s.LogTime = zerolog.TimeFormatUnix
	case "unixms":
		s.LogTime = zerolog.TimeFormatUnixMs
	case "unixmicro":
		s.LogTime = zerolog.TimeFormatUnixMicro
	default:
		log.Warn().Msgf("unknown log time %s", logTime)
	}

	s.LogJson = viper.GetBool("log.json")
	s.LogNocolor = viper.GetBool("log.nocolor")
	s.LogDir = viper.GetString("log.dir")

	if viper.GetBool("debug") && s.LogLevel != zerolog.TraceLevel {
		s.LogLevel = zerolog.DebugLevel
	}

	// support for NO_COLOR env variable: https://no-color.org/
	if os.Getenv("NO_COLOR") != "" {
		s.LogNocolor = true
	}
}

func (s *Root) SetV2() {
	if viper.IsSet("logs") {
		if viper.GetBool("logs") {
			logs := filepath.Join(".", "logs")
			if runtime.GOOS == "linux" {
				logs = "/var/log/neko"
			}
			s.LogDir = logs
		} else {
			s.LogDir = ""
		}
		log.Warn().Msg("you are using v2 configuration 'NEKO_LOGS' which is deprecated, please use 'NEKO_LOG_DIR=/path/to/logs' instead")
	}
}
