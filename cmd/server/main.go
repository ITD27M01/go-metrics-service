package main

import (
	"context"
	"fmt"
	"github.com/itd27m01/go-metrics-service/internal/pkg/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	metricsServer := server.MetricsServer{
		Cfg: server.Config{
			ServerPort:    "8080",
			ServerAddress: "0.0.0.0",
		}}

	go metricsServer.StartListener(context.Background())

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	<-signalChanel

	fmt.Println("Stop signal received, graceful shutdown the server...")
	metricsServer.StopListener()
}
