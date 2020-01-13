package main

import (
	"fmt"

	"n.eko.moe/neko"
	"n.eko.moe/neko/cmd"
	"n.eko.moe/neko/internal/utils"

	"github.com/rs/zerolog/log"
)

func main() {
	fmt.Print(utils.Colorf(utils.Header, "server", neko.Service.Version))
	if err := cmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("Failed to execute command")
	}
}
