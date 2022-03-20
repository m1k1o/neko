package main

import (
	"fmt"

	"github.com/rs/zerolog/log"

	neko "gitlab.com/demodesk/neko/server"
	"gitlab.com/demodesk/neko/server/cmd"
	"gitlab.com/demodesk/neko/server/pkg/utils"
)

func main() {
	fmt.Print(utils.Colorf(neko.Header, "server", neko.Version))
	if err := cmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("failed to execute command")
	}
}
