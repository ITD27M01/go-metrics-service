package server

import (
	"context"
	"log"

	"github.com/itd27m01/go-metrics-service/internal/preserver"
	"github.com/itd27m01/go-metrics-service/internal/repository"
)

func runStore(ctx context.Context, config *Config) func() error {
	switch {
	case config.DatabaseDSN != "":
		metricsStore, err := repository.NewDBStore(config.DatabaseDSN)
		if err != nil {
			log.Fatalf("Couldn't connect to database: %q", err)
		}

		config.MetricsStore = metricsStore

		log.Println("Using Database storage")

		return func() error {
			return metricsStore.Close()
		}
	case config.StoreFilePath != "":
		syncChannel := make(chan struct{}, 1)
		metricsStore, err := repository.NewFileStore(config.StoreFilePath, syncChannel)
		if err != nil {
			log.Fatalf("Failed to make file storage: %q", err)
		}
		config.MetricsStore = metricsStore

		log.Println("Using file storage")

		metricsPreserver := preserver.NewPreserver(metricsStore, config.StoreInterval, syncChannel)

		if config.Restore && metricsStore.LoadMetrics() != nil {
			log.Println("Filed to load metrics from store")
		}

		preserverContext, preserverCancel := context.WithCancel(ctx)

		go metricsPreserver.RunPreserver(preserverContext)

		return func() error {
			var err error
			if err = metricsStore.SaveMetrics(); err != nil {
				log.Printf("Something went wrong during metrics preserve %q", err)
			}

			if err = metricsStore.Close(); err != nil {
				log.Printf("Something went wrong during file close %q", err)
			}
			preserverCancel()

			return err
		}
	default:
		log.Println("Using memory storage")
		config.MetricsStore = repository.NewInMemoryStore()

		return func() error {
			return nil
		}
	}
}
