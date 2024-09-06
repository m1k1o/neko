package main

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko"
	"github.com/demodesk/neko/cmd"
	"github.com/demodesk/neko/pkg/utils"
)

func main() {
	fmt.Print(utils.Colorf(neko.Header, "server", neko.Version))
	if err := cmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("failed to execute command")
	}
}
