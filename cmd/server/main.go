package main

import (
	"context"

	"github.com/caarlos0/env/v6"
	"github.com/itd27m01/go-metrics-service/cmd/server/cmd"
	"github.com/itd27m01/go-metrics-service/internal/pkg/logging"
	"github.com/itd27m01/go-metrics-service/internal/pkg/logging/log"
	"github.com/itd27m01/go-metrics-service/internal/server"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse command line arguments")
	}

	logging.LogLevel(cmd.LogLevel)

	metricsServerConfig := server.Config{
		ServerAddress: cmd.ServerAddress,
		StoreInterval: cmd.StoreInterval,
		Restore:       cmd.Restore,
		StoreFilePath: cmd.StoreFilePath,
		DatabaseDSN:   cmd.DatabaseDSN,
		SignKey:       cmd.SignKey,
		LogLevel:      cmd.LogLevel,
	}
	if err := env.Parse(&metricsServerConfig); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse environment variables")
	}

	metricsServer := server.MetricsServer{Cfg: &metricsServerConfig}

	metricsServer.Start(context.Background())
}
