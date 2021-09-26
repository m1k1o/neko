package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"demodesk/neko"
)

func Execute() error {
	return root.Execute()
}

var root = &cobra.Command{
	Use:     "neko",
	Short:   "neko streaming server",
	Long:    `neko streaming server`,
	Version: neko.Service.Version.String(),
}

func init() {
	cobra.OnInitialize(func() {
		//////
		// logs
		//////
		console := zerolog.ConsoleWriter{Out: os.Stdout}
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		zerolog.SetGlobalLevel(zerolog.InfoLevel)

		logs := viper.GetBool("logs")
		if !logs {
			log.Logger = log.Output(console)
		} else {
			logsPath := filepath.Join(".", "logs")
			if runtime.GOOS == "linux" {
				logsPath = "/var/log/neko"
			}

			if _, err := os.Stat(logsPath); os.IsNotExist(err) {
				_ = os.Mkdir(logsPath, os.ModePerm)
			}

			latest := filepath.Join(logsPath, "neko-latest.log")
			if _, err := os.Stat(latest); err == nil {
				err = os.Rename(latest, filepath.Join(logsPath, "neko."+time.Now().Format("2006-01-02T15-04-05Z07-00")+".log"))
				if err != nil {
					log.Panic().Err(err).Msg("failed to rotate log file")
				}
			}

			logf, err := os.OpenFile(latest, os.O_RDWR|os.O_CREATE, 0666)
			if err != nil {
				log.Panic().Err(err).Msg("failed to create log file")
			}

			logger := diode.NewWriter(logf, 1000, 10*time.Millisecond, func(missed int) {
				fmt.Printf("logger dropped %d messages", missed)
			})

			log.Logger = log.Output(io.MultiWriter(console, logger))
		}

		//////
		// configs
		//////
		config := viper.GetString("config") // Use config file from the flag.
		if config == "" {
			config = os.Getenv("NEKO_CONFIG") // Use config file from the environment variable.
		}

		if config != "" {
			viper.SetConfigFile(config)
		} else {
			if runtime.GOOS == "linux" {
				viper.AddConfigPath("/etc/neko/")
			}

			viper.AddConfigPath(".")
			viper.SetConfigName("neko")
		}

		viper.SetEnvPrefix("NEKO")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv() // read in environment variables that match

		err := viper.ReadInConfig()
		if err != nil && config != "" {
			log.Err(err)
		}

		// get full config file path
		config = viper.ConfigFileUsed()

		//////
		// debug
		//////
		debug := viper.GetBool("debug")
		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		logger := log.With().
			Bool("debug", debug).
			Bool("logs", logs).
			Str("config", config).
			Logger()

		if config == "" {
			logger.Warn().Msg("preflight complete without config file")
		} else {
			if _, err := os.Stat(config); os.IsNotExist(err) {
				logger.Err(err).Msg("preflight complete with nonexistent config file")
			} else {
				logger.Info().Msg("preflight complete with config file")
			}
		}

		neko.Service.Configs.Root.Set()
	})

	if err := neko.Service.Configs.Root.Init(root); err != nil {
		log.Panic().Err(err).Msg("unable to run root command")
	}

	root.SetVersionTemplate(neko.Service.Version.Details())
}
