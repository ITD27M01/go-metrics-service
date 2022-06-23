package server

import (
	"context"

	"github.com/itd27m01/go-metrics-service/internal/preserver"
	"github.com/itd27m01/go-metrics-service/internal/repository"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

// initStore creates storage repository for metrics
func initStore(ctx context.Context, config *Config) func() error {
	switch {
	case config.DatabaseDSN != "":
		metricsStore, err := repository.NewDBStore(config.DatabaseDSN)
		if err != nil {
			log.Fatal().Msgf("Couldn't connect to database: %q", err)
		}

		config.MetricsStore = metricsStore

		log.Info().Msg("Using Database storage")

		return func() error {
			return metricsStore.Close()
		}
	case config.StoreFilePath != "":
		syncChannel := make(chan struct{}, 1)
		metricsStore, err := repository.NewFileStore(config.StoreFilePath, syncChannel)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to make file storage")
		}
		config.MetricsStore = metricsStore

		log.Info().Msg("Using file storage")

		metricsPreserver := preserver.NewPreserver(metricsStore, config.StoreInterval, syncChannel)

		if config.Restore && metricsStore.LoadMetrics() != nil {
			log.Error().Msg("Filed to load metrics from store")
		}

		preserverContext, preserverCancel := context.WithCancel(ctx)

		go metricsPreserver.RunPreserver(preserverContext)

		return func() error {
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
		config.MetricsStore = repository.NewInMemoryStore()

		return func() error {
			return nil
		}
	}
}
