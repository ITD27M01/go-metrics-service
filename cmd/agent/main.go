package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/itd27m01/go-metrics-service/internal/pkg/repository"
	"github.com/itd27m01/go-metrics-service/internal/pkg/workers"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	serverTimeout  = 1 * time.Second
)

func main() {
	mtr := repository.NewInMemoryStore()

	pollWorkerConfig := workers.PollerConfig{
		PollInterval: pollInterval,
	}
	err := env.Parse(&pollWorkerConfig)
	if err != nil {
		log.Fatal(err)
	}

	pollWorker := workers.PollerWorker{Cfg: pollWorkerConfig}
	pollContext, cancelCollector := context.WithCancel(context.Background())
	go pollWorker.Run(pollContext, mtr)

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

	reportWorker := workers.ReportWorker{Cfg: reportWorkerConfig}

	reportContext, cancelReporter := context.WithCancel(context.Background())
	go reportWorker.Run(reportContext, mtr)

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	log.Printf("%v signal received, stopping collector worker", <-signalChanel)
	cancelCollector()

	log.Println("...stopping reporter worker")
	cancelReporter()

	log.Println("All workers are stopped")
}
