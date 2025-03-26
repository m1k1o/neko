package main

import (
	"fmt"

	"github.com/rs/zerolog/log"

	neko "github.com/m1k1o/neko/server"
	"github.com/m1k1o/neko/server/cmd"
	"github.com/m1k1o/neko/server/pkg/utils"
)

func main() {
	fmt.Print(utils.Colorf(neko.Header, "server", neko.Version))
	if err := cmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("failed to execute command")
	}
}
