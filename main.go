package main

import (
	"flag"

	"go.uber.org/zap"

	"github.com/levente-simon/terraform-elastic-backend/server"
)

func main() {
	var configFilePath string

	// Initialize Zap logger
	logger, _ := zap.NewProduction()
	defer logger.Sync() // Ensure logs are flushed before exiting

	// Parse command-line flags
	flag.StringVar(&configFilePath, "config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// Log the configuration file path being used
	logger.Info("Using configuration file", zap.String("path", configFilePath))

	// Start the HTTP server
	if err := server.ServeHttp(configFilePath, logger); err != nil {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}
