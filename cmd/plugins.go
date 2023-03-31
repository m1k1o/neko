package cmd

import (
	"encoding/json"
	"os"

	"github.com/demodesk/neko/internal/config"
	"github.com/demodesk/neko/internal/plugins"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	command := &cobra.Command{
		Use:   "plugins [directory]",
		Short: "load, verify and list plugins",
		Long:  `load, verify and list plugins`,
		Run:   pluginsCmd,
		Args:  cobra.MaximumNArgs(1),
	}
	root.AddCommand(command)
}

func pluginsCmd(cmd *cobra.Command, args []string) {
	pluginDir := "/etc/neko/plugins"
	if len(args) > 0 {
		pluginDir = args[0]
	}
	log.Info().Str("dir", pluginDir).Msg("plugins directory")

	plugs := plugins.New(&config.Plugins{
		Enabled:  true,
		Required: true,
		Dir:      pluginDir,
	})

	meta := plugs.Metadata()
	if len(meta) == 0 {
		log.Fatal().Msg("no plugins found")
	}

	// marshal indent to stdout
	dec := json.NewEncoder(os.Stdout)
	dec.SetIndent("", "  ")
	err := dec.Encode(meta)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to marshal metadata")
	}
}
