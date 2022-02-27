package server

import (
	"log"

	"github.com/itd27m01/go-metrics-service/internal/preserver"
	"github.com/itd27m01/go-metrics-service/internal/repository"
)

func initStore(config *Config) *preserver.Preserver {
	syncChannel := make(chan struct{}, 1)

	if config.StoreFilePath == "" {
		config.MetricsStore = repository.NewInMemoryStore()
	} else {
		fileStore, err := repository.NewFileStore(config.StoreFilePath, syncChannel)
		if err != nil {
			log.Fatalf("Failed to make file storage: %q", err)
		}

		config.MetricsStore = fileStore
	}

	return preserver.NewPreserver(config.MetricsStore, config.StoreInterval, syncChannel)
}
