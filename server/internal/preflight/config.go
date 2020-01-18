package preflight

import (
	"runtime"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func Config(name string) {
	config := viper.GetString("config")

	if config != "" {
		viper.SetConfigFile(config) // Use config file from the flag.
	} else {
		if runtime.GOOS == "linux" {
			viper.AddConfigPath("/etc/neko/")
		}

		viper.AddConfigPath(".")
		viper.SetConfigName(name)
	}

	viper.SetEnvPrefix("NEKO")
	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Error().Err(err)
		}
		if config != "" {
			log.Error().Err(err)
		}
	}

	file := viper.ConfigFileUsed()
	logger := log.With().
		Bool("debug", viper.GetBool("debug")).
		Str("logging", viper.GetString("logs")).
		Str("config", file).
		Logger()

	if file == "" {
		logger.Warn().Msg("preflight complete without config file")
	} else {
		logger.Info().Msg("preflight complete")
	}
}
