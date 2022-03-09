package server

import (
	"context"
	"log"

	"github.com/itd27m01/go-metrics-service/internal/repository"
)

func initStore(ctx context.Context, config *Config) chan struct{} {
	syncChannel := make(chan struct{}, 1)

	switch {
	case config.DatabaseDSN != "":
		log.Println("Using Database storage")
		metricsStore, err := repository.NewDBStore(ctx, config.DatabaseDSN, syncChannel)
		if err != nil {
			log.Fatalf("Couldn't connect to database: %q", err)
		}
		config.MetricsStore = metricsStore
	case config.StoreFilePath != "":
		log.Println("Using file storage")
		fileStore, err := repository.NewFileStore(config.StoreFilePath, syncChannel)
		if err != nil {
			log.Fatalf("Failed to make file storage: %q", err)
		}
		config.MetricsStore = fileStore
	default:
		log.Println("Using memory storage")
		config.MetricsStore = repository.NewInMemoryStore()
	}

	return syncChannel
}
