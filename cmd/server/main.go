package main

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/caarlos0/env/v6"
	"github.com/itd27m01/go-metrics-service/cmd/server/cmd"
	"github.com/itd27m01/go-metrics-service/internal/server"
	"github.com/rs/zerolog"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse command line arguments")
	}

	switch cmd.LogLevel {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARNING":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}

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
		log.Fatal().Err(err).Msg("")
	}

	metricsServer := server.MetricsServer{Cfg: &metricsServerConfig}

	metricsServer.Start(context.Background())
}
