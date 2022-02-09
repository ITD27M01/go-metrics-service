package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	"github.com/itd27m01/go-metrics-service/internal/pkg/workers"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	serverTimeout  = 1 * time.Second
)

func main() {
	mtr := metrics.NewMetrics()

	pollWorker := workers.PollerWorker{Cfg: workers.PollerConfig{PollInterval: pollInterval}}
	pollContext, cancelCollector := context.WithCancel(context.Background())
	go pollWorker.Run(pollContext, mtr)

	reportWorker := workers.ReportWorker{
		Cfg: workers.ReporterConfig{
			ServerURL:      "http://127.0.0.1:8080/update",
			ServerTimeout:  serverTimeout,
			ReportInterval: reportInterval,
		}}

	reportContext, cancelReporter := context.WithCancel(context.Background())
	go reportWorker.Run(reportContext, mtr)

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	<-signalChanel

	log.Println("Stop signal received, stopping collector worker...")
	cancelCollector()

	log.Println("...stopping reporter worker")
	cancelReporter()

	log.Println("All workers are stopped")
}
