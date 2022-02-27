package server

import (
	"context"
	"log"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/repository"
)

func initMetricsStore(config *Config) {
	if config.StoreFilePath == "" {
		config.MetricsStore = repository.NewInMemoryStore()

		return
	}

	fileStore, err := repository.NewFileStore(config.StoreFilePath)
	if err != nil {
		log.Printf("Failed to make file storage: %q", err)
		config.MetricsStore = repository.NewInMemoryStore()
	} else {
		config.MetricsStore = fileStore
	}
}

func runPreserver(ctx context.Context, store repository.Store, restore bool, storeInterval time.Duration) {
	if restore {
		err := store.LoadMetrics()
		if err != nil {
			log.Printf("Filed to load metrics from file: %q", err)
		}
	}

	store.RunPreserver(ctx, storeInterval)
	log.Println("Preserver exited...")
}
