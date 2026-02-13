// Package main for stackrox-mcp command.
package main

import (
	"context"
	"flag"

	"github.com/stackrox/stackrox-mcp/internal/app"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/logging"
)

func main() {
	logging.SetupLogging()

	configPath := flag.String("config", "", "Path to configuration file (optional)")

	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logging.Fatal("Failed to load configuration", err)
	}

	ctx := context.Background()

	if err := app.Run(ctx, cfg, nil, nil); err != nil {
		logging.Fatal("Server error", err)
	}
}

