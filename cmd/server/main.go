package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	"github.com/itd27m01/go-metrics-service/internal/pkg/server"
)

func main() {
	metricsServer := server.MetricsServer{
		Cfg: server.Config{
			ServerPort:    "8080",
			ServerAddress: "0.0.0.0",
			MetricsData:   metrics.NewMetrics(),
		}}

	go metricsServer.StartListener(context.Background())

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	<-signalChanel

	log.Println("Stop signal received, graceful shutdown the server...")
	metricsServer.StopListener()
}
