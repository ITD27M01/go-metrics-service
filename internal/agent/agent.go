package agent

import (
	"context"
	"os/signal"
	"sync"
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
func Start(parent context.Context, config *AgentConfig) {
	ctx, stop := signal.NotifyContext(parent,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	wg := sync.WaitGroup{}

	metricsStore := repository.NewInMemoryStore()

	pollWorker := PollerWorker{Cfg: &config.PollerConfig}

	wg.Add(1)
	go func() {
		defer wg.Done()

		pollWorker.RunMemStats(ctx, metricsStore)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		pollWorker.RunPsStats(ctx, metricsStore)
	}()

	reportWorker := ReportWorker{Cfg: &config.ReporterConfig}

	wg.Add(1)
	go func() {
		defer wg.Done()

		reportWorker.Run(ctx, metricsStore)
	}()

	wg.Wait()
	log.Info().Msg("All workers are stopped")
}
