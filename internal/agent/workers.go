package agent

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/itd27m01/go-metrics-service/internal/pkg/logging/log"
	"github.com/itd27m01/go-metrics-service/internal/repository"
)

func Start(ctx context.Context, pollWorkerConfig PollerConfig, reportWorkerConfig ReporterConfig) {
	metricsStore := repository.NewInMemoryStore()

	pollWorker := PollerWorker{Cfg: pollWorkerConfig}
	pollContext, cancelCollector := context.WithCancel(ctx)
	go pollWorker.Run(pollContext, metricsStore)

	reportWorker := ReportWorker{Cfg: reportWorkerConfig}

	reportContext, cancelReporter := context.WithCancel(ctx)
	go reportWorker.Run(reportContext, metricsStore)

	log.Info().Msgf("%v signal received, stopping collector worker", <-getSignalChannel())
	cancelCollector()

	log.Info().Msg("...stopping reporter worker")
	cancelReporter()

	log.Info().Msg("All workers are stopped")
}

func getSignalChannel() chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	return signalChannel
}
