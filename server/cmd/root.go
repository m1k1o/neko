package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"n.eko.moe/neko"
	"n.eko.moe/neko/internal/preflight"
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
		preflight.Logs("neko")
		preflight.Config("neko")
		neko.Service.Root.Set()
	})

	if err := neko.Service.Root.Init(root); err != nil {
		log.Panic().Err(err).Msg("unable to run root command")
	}

	root.SetVersionTemplate(neko.Service.Version.Details())
}
