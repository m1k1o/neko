package main

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"demodesk/neko"
	"demodesk/neko/cmd"
	"demodesk/neko/internal/utils"
)

func main() {
	fmt.Print(utils.Colorf(neko.Header, "server", neko.Service.Version))
	if err := cmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("failed to execute command")
	}
}
