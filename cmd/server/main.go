package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	"github.com/itd27m01/go-metrics-service/internal/pkg/server"
)

func main() {
	metricsServerConfig := server.Config{
		ServerAddress: "0.0.0.0:8080",
		MetricsStore:  metrics.NewInMemoryStore(),
	}
	err := env.Parse(&metricsServerConfig)
	if err != nil {
		log.Fatal(err)
	}

	metricsServer := server.MetricsServer{Cfg: metricsServerConfig}

	go metricsServer.StartListener(context.Background())
	log.Printf("Start listener on %s", metricsServer.Cfg.ServerAddress)

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	log.Printf("%v signal received, graceful shutdown the server", <-signalChanel)
	metricsServer.StopListener()
}
