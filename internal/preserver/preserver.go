package preserver

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/itd27m01/go-metrics-service/internal/repository"
)

type Preserver struct {
	store         *repository.FileStore
	storeInterval time.Duration
	syncChannel   chan struct{}
}

func NewPreserver(store *repository.FileStore, storeInterval time.Duration, syncChannel chan struct{}) *Preserver {
	p := Preserver{
		store:         store,
		storeInterval: storeInterval,
		syncChannel:   syncChannel,
	}

	return &p
}

func (p *Preserver) RunPreserver(ctx context.Context) {
	log.Info().Msg("Run preserver for metrics")

	pollTicker := new(time.Ticker)
	if p.storeInterval > 0 {
		pollTicker = time.NewTicker(p.storeInterval)

		log.Info().Msgf("Dump metrics every %s", p.storeInterval)
	}
	defer pollTicker.Stop()

	var err error
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
				log.Error().Err(err).Msgf("Something went wrong during metrics preserve")
			}

			return
		}

		if err != nil {
			log.Error().Err(err).Msgf("Something went wrong during metrics preserve")
		}
	}
}
