package main

import (
	"context"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/itd27m01/go-metrics-service/internal/pkg/workers"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	serverTimeout  = 1 * time.Second
)

func main() {
	pollWorkerConfig := workers.PollerConfig{
		PollInterval: pollInterval,
	}
	err := env.Parse(&pollWorkerConfig)
	if err != nil {
		log.Fatal(err)
	}

	reportWorkerConfig := workers.ReporterConfig{
		ServerScheme:   "http",
		ServerAddress:  "127.0.0.1:8080",
		ServerPath:     "/update/",
		ServerTimeout:  serverTimeout,
		ReportInterval: reportInterval,
	}
	err = env.Parse(&reportWorkerConfig)
	if err != nil {
		log.Fatal(err)
	}

	workers.Start(context.Background(), pollWorkerConfig, reportWorkerConfig)
}
