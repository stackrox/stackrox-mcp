package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func setupLogging() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func main() {
	setupLogging()

	log.Info().Msg("Starting Stackrox MCP server")
}
