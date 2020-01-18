package preflight

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func Logs(name string) {
	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if viper.GetBool("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	console := zerolog.ConsoleWriter{Out: os.Stdout}

	if !viper.GetBool("logs") {
		log.Logger = log.Output(console)
	} else {

		logs := filepath.Join(".", "logs")
		if runtime.GOOS == "linux" {
			logs = "/var/log/neko"
		}

		if _, err := os.Stat(logs); os.IsNotExist(err) {
			os.Mkdir(logs, os.ModePerm)
		}

		latest := filepath.Join(logs, name+"-latest.log")
		_, err := os.Stat(latest)
		if err == nil {
			err = os.Rename(latest, filepath.Join(logs, "neko."+time.Now().Format("2006-01-02T15-04-05Z07-00")+".log"))
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
}
