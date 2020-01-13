package cmd

import (
	"fmt"

	"n.eko.moe/neko"
	"n.eko.moe/neko/internal/preflight"

	"github.com/spf13/cobra"
)

func Execute() error {
	return root.Execute()
}

var root = &cobra.Command{
	Use:     "neko",
	Short:   "",
	Long:    ``,
	Version: neko.Service.Version.String(),
}

func init() {
	cobra.OnInitialize(func() {
		preflight.Logs("neko")
		preflight.Config("neko")
		neko.Service.Root.Set()
	})

	if err := neko.Service.Root.Init(root); err != nil {
		neko.Service.Logger.Panic().Err(err).Msg("Unable to run command")
	}

	root.SetVersionTemplate(fmt.Sprintf("Version: %s\n", neko.Service.Version))
}
