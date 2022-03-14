package main

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/caarlos0/env/v6"
	"github.com/itd27m01/go-metrics-service/cmd/agent/cmd"
	"github.com/itd27m01/go-metrics-service/internal/agent"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal().Msgf("Failed to parse command line arguments: %v", err)
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

	pollWorkerConfig := agent.PollerConfig{
		PollInterval: cmd.PollInterval,
	}
	if err := env.Parse(&pollWorkerConfig); err != nil {
		log.Fatal().Msgf("%v", err)
	}

	reportWorkerConfig := agent.ReporterConfig{
		ServerScheme:   "http",
		ServerAddress:  cmd.ServerAddress,
		ServerPath:     "/update/",
		ServerTimeout:  cmd.ServerTimeout,
		ReportInterval: cmd.ReportInterval,
		SignKey:        cmd.SignKey,
	}
	if err := env.Parse(&reportWorkerConfig); err != nil {
		log.Fatal().Msgf("%v", err)
	}

	agent.Start(context.Background(), pollWorkerConfig, reportWorkerConfig)
}
