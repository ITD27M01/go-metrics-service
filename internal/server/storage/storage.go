package storage

import (
	"context"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/preserver"
	"github.com/itd27m01/go-metrics-service/internal/repository"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

// Config collects configuration for metrics storage
type Config struct {
	DatabaseDSN   string        `yaml:"database_dsn" env:"DATABASE_DSN"`
	StoreFilePath string        `yaml:"store_file_path" env:"STORE_FILE"`
	StoreInterval time.Duration `yaml:"store_interval" env:"STORE_INTERVAL"`
	Restore       bool          `yaml:"restore" env:"RESTORE"`
}

// StartMetricsStorage starts storage repository for metrics
func StartMetricsStorage(ctx context.Context, config *Config) (repository.Store, func() error) {
	switch {
	case config.DatabaseDSN != "":
		metricsStore, err := repository.NewDBStore(config.DatabaseDSN)
		if err != nil {
			log.Fatal().Msgf("Couldn't connect to database: %q", err)
		}

		log.Info().Msg("Using Database storage")

		return metricsStore, func() error {
			return metricsStore.Close()
		}
	case config.StoreFilePath != "":
		syncChannel := make(chan struct{}, 1)
		metricsStore, err := repository.NewFileStore(config.StoreFilePath, syncChannel)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to make file storage")
		}

		log.Info().Msg("Using file storage")

		metricsPreserver := preserver.NewPreserver(metricsStore, config.StoreInterval, syncChannel)

		if config.Restore && metricsStore.LoadMetrics() != nil {
			log.Error().Msg("Filed to load metrics from store")
		}

		preserverContext, preserverCancel := context.WithCancel(ctx)

		go metricsPreserver.RunPreserver(preserverContext)

		return metricsStore, func() error {
			var err error
			if err = metricsStore.SaveMetrics(); err != nil {
				log.Error().Err(err).Msg("Something went wrong during metrics preserve")
			}

			if err = metricsStore.Close(); err != nil {
				log.Error().Err(err).Msg("Something went wrong during file close")
			}
			preserverCancel()

			return err
		}
	default:
		log.Info().Msg("Using memory storage")

		return repository.NewInMemoryStore(), func() error {
			return nil
		}
	}
}
