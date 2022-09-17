package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"m1k1o/neko"
	"m1k1o/neko/internal/config"
)

func init() {
	command := &cobra.Command{
		Use:   "serve",
		Short: "serve neko streaming server",
		Long:  `serve neko streaming server`,
		Run:   neko.Service.ServeCommand,
	}

	configs := []config.Config{
		neko.Service.Server,
		neko.Service.WebRTC,
		neko.Service.Capture,
		neko.Service.Desktop,
		neko.Service.WebSocket,
	}

	cobra.OnInitialize(func() {
		for _, cfg := range configs {
			cfg.Set()
		}
		neko.Service.Preflight()
	})

	for _, cfg := range configs {
		if err := cfg.Init(command); err != nil {
			log.Panic().Err(err).Msg("unable to run serve command")
		}
	}

	root.AddCommand(command)
}
