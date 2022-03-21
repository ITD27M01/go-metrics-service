package main

import (
	"context"

	"github.com/caarlos0/env/v6"
	"github.com/itd27m01/go-metrics-service/cmd/agent/cmd"
	"github.com/itd27m01/go-metrics-service/internal/agent"
	"github.com/itd27m01/go-metrics-service/internal/pkg/logging"
	"github.com/itd27m01/go-metrics-service/internal/pkg/logging/log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal().Msgf("Failed to parse command line arguments: %v", err)
	}

	logging.LogLevel(cmd.LogLevel)

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
		log.Fatal().Err(err).Msg("Failed to parse environment variables")
	}

	agent.Start(context.Background(), pollWorkerConfig, reportWorkerConfig)
}
