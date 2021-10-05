package main

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"m1k1o/neko"
	"m1k1o/neko/cmd"
	"m1k1o/neko/internal/utils"
)

func main() {
	fmt.Print(utils.Colorf(neko.Header, "server", neko.Service.Version))
	if err := cmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("failed to execute command")
	}
}
