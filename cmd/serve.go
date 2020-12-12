package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"demodesk/neko"
	"demodesk/neko/modules"
	"demodesk/neko/internal/config"
)

func init() {
	command := &cobra.Command{
		Use:   "serve",
		Short: "serve neko streaming server",
		Long:  `serve neko streaming server`,
		Run:   neko.Service.ServeCommand,
	}

	configs := append([]config.Config{
		neko.Service.Configs.Desktop,
		neko.Service.Configs.Capture,
		neko.Service.Configs.WebRTC,
		neko.Service.Configs.Session,
		neko.Service.Configs.Server,
	}, modules.ConfigsList()...)

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
