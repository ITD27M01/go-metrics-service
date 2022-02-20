package main

import (
	"context"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/itd27m01/go-metrics-service/cmd/agent/cmd"
	"github.com/itd27m01/go-metrics-service/internal/pkg/workers"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Printf("Failed to parse command line arguments: %q", err)

		return
	}

	pollWorkerConfig := workers.PollerConfig{
		PollInterval: cmd.PollInterval,
	}
	err = env.Parse(&pollWorkerConfig)
	if err != nil {
		log.Fatal(err)
	}

	reportWorkerConfig := workers.ReporterConfig{
		ServerScheme:   "http",
		ServerAddress:  cmd.ServerAddress,
		ServerPath:     "/update/",
		ServerTimeout:  cmd.ServerTimeout,
		ReportInterval: cmd.ReportInterval,
	}
	err = env.Parse(&reportWorkerConfig)
	if err != nil {
		log.Fatal(err)
	}

	workers.Start(context.Background(), pollWorkerConfig, reportWorkerConfig)
}
