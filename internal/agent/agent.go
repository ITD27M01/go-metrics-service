package agent

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/itd27m01/go-metrics-service/internal/repository"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

type AgentConfig struct {
	PollerConfig   PollerConfig   `yaml:"poller"`
	ReporterConfig ReporterConfig `yaml:"reporter"`
	LogLevel       string         `yaml:"log_level" env:"LOG_LEVEL"`
}

// Start starts poller and reporter agent's workers
func Start(ctx context.Context, config *AgentConfig) {
	metricsStore := repository.NewInMemoryStore()

	pollWorker := PollerWorker{Cfg: &config.PollerConfig}
	pollContext, cancelPoller := context.WithCancel(ctx)
	go pollWorker.RunMemStats(pollContext, metricsStore)
	go pollWorker.RunPsStats(pollContext, metricsStore)

	reportWorker := ReportWorker{Cfg: &config.ReporterConfig}

	reportContext, cancelReporter := context.WithCancel(ctx)
	go reportWorker.Run(reportContext, metricsStore)

	log.Info().Msgf("%v signal received, stopping collector worker", <-getSignalChannel())
	cancelPoller()

	log.Info().Msg("...stopping reporter worker")
	cancelReporter()

	log.Info().Msg("All workers are stopped")
}

// getSignalChannel returns a channel for waiting and Cntrl-C signal
func getSignalChannel() chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	return signalChannel
}
