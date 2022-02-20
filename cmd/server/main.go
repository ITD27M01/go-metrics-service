package main

import (
	"context"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/itd27m01/go-metrics-service/cmd/server/cmd"
	"github.com/itd27m01/go-metrics-service/internal/pkg/server"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Printf("Failed to parse command line arguments: %q", err)

		return
	}

	metricsServerConfig := server.Config{
		ServerAddress: cmd.ServerAddress,
		StoreInterval: cmd.StoreInterval,
		Restore:       cmd.Restore,
		StoreFile:     cmd.StoreFile,
	}
	err = env.Parse(&metricsServerConfig)
	if err != nil {
		log.Fatal(err)
	}

	metricsServer := server.MetricsServer{Cfg: &metricsServerConfig}

	metricsServer.Start(context.Background())
}
