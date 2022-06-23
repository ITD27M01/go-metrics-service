package preserver

import (
	"context"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/repository"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

// Preserver defines worker to preserve the metrics in file store
type Preserver struct {
	store         *repository.FileStore
	storeInterval time.Duration
	syncChannel   chan struct{}
}

// NewPreserver creates preserver
func NewPreserver(store *repository.FileStore, storeInterval time.Duration, syncChannel chan struct{}) *Preserver {
	p := Preserver{
		store:         store,
		storeInterval: storeInterval,
		syncChannel:   syncChannel,
	}

	return &p
}

// RunPreserver runs preserver worker
func (p *Preserver) RunPreserver(ctx context.Context) {
	log.Info().Msg("Run preserver for metrics")

	pollTicker := new(time.Ticker)
	if p.storeInterval > 0 {
		pollTicker = time.NewTicker(p.storeInterval)

		log.Info().Msgf("Dump metrics every %s", p.storeInterval)
	}
	defer pollTicker.Stop()

	var err error
	const errMessage = "Something went wrong during metrics preserve"
	for {
		select {
		case <-pollTicker.C:
			err = p.store.SaveMetrics()
		case <-p.syncChannel:
			if p.storeInterval == 0 {
				err = p.store.SaveMetrics()
			}
		case <-ctx.Done():
			err = p.store.SaveMetrics()
			if err != nil {
				log.Error().Err(err).Msgf(errMessage)
			}

			return
		}

		if err != nil {
			log.Error().Err(err).Msgf(errMessage)
		}
	}
}
