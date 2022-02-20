package main

import (
	"context"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/itd27m01/go-metrics-service/internal/pkg/server"
)

func main() {
	metricsServerConfig := server.Config{}
	err := env.Parse(&metricsServerConfig)
	if err != nil {
		log.Fatal(err)
	}

	metricsServer := server.MetricsServer{Cfg: &metricsServerConfig}

	metricsServer.Start(context.Background())
}
