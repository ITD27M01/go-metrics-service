package server

import (
	"context"
	"github.com/itd27m01/go-metrics-service/internal/preserver"
	"github.com/itd27m01/go-metrics-service/internal/repository"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

// startServerStorage starts storage repository for metrics
func startServerStorage(ctx context.Context, server *MetricsServer) func() error {
	switch {
	case server.Cfg.DatabaseDSN != "":
		metricsStore, err := repository.NewDBStore(server.Cfg.DatabaseDSN)
		if err != nil {
			log.Fatal().Msgf("Couldn't connect to database: %q", err)
		}

		server.metricsStore = metricsStore

		log.Info().Msg("Using Database storage")

		return func() error {
			return metricsStore.Close()
		}
	case server.Cfg.StoreFilePath != "":
		syncChannel := make(chan struct{}, 1)
		metricsStore, err := repository.NewFileStore(server.Cfg.StoreFilePath, syncChannel)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to make file storage")
		}
		server.metricsStore = metricsStore

		log.Info().Msg("Using file storage")

		metricsPreserver := preserver.NewPreserver(metricsStore, server.Cfg.StoreInterval, syncChannel)

		if server.Cfg.Restore && metricsStore.LoadMetrics() != nil {
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
		server.metricsStore = repository.NewInMemoryStore()

		return func() error {
			return nil
		}
	}
}
