package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"n.eko.moe/neko"
	"n.eko.moe/neko/internal/config"
)

func init() {
	command := &cobra.Command{
		Use:   "serve",
		Short: "",
		Long:  ``,
		Run:   neko.Service.ServeCommand,
	}

	configs := []config.Config{
		neko.Service.Serve,
	}

	cobra.OnInitialize(func() {
		for _, cfg := range configs {
			cfg.Set()
		}
		neko.Service.Preflight()
	})

	for _, cfg := range configs {
		if err := cfg.Init(command); err != nil {
			log.Panic().Err(err).Msg("Unable to run command")
		}
	}

	root.AddCommand(command)
}
