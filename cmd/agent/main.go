package main

import (
	"context"
	"github.com/caarlos0/env/v6"
	"github.com/itd27m01/go-metrics-service/cmd/agent/cmd"
	"github.com/itd27m01/go-metrics-service/internal/agent"
	"github.com/itd27m01/go-metrics-service/internal/greetings"
	"github.com/itd27m01/go-metrics-service/pkg/logging"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
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
		CryptoKey:      cmd.CryptoKey,
		SignKey:        cmd.SignKey,
	}
	if err := env.Parse(&reportWorkerConfig); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse environment variables")
	}

	if err := greetings.Print(buildVersion, buildDate, buildCommit); err != nil {
		log.Fatal().Err(err).Msg("Failed to start agent, failed to print greetings")
	}
	agent.Start(context.Background(), pollWorkerConfig, reportWorkerConfig)
}
