package server

import (
	"context"
	"log"

	"github.com/itd27m01/go-metrics-service/internal/pkg/repository"
)

func initMetricsStore(config *Config) {
	if config.StoreFile == "" {
		config.MetricsStore = repository.NewInMemoryStore()

		return
	}

	fileStore, err := repository.NewFileStore(config.StoreFile, config.StoreInterval)
	if err != nil {
		log.Printf("Failed to make file storage: %q", err)
		config.MetricsStore = repository.NewInMemoryStore()
	} else {
		config.MetricsStore = fileStore
	}
}

func runPreserver(ctx context.Context, store repository.Store, restore bool) {
	if restore {
		err := store.LoadMetrics()
		if err != nil {
			log.Printf("Filed to load metrics from file: %q", err)
		}
	}

	store.RunPreserver(ctx)
	log.Println("Preserver exited...")
}
