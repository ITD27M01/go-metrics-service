package workers

import (
	"context"
	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	"time"
)

type PollerConfig struct {
	PollInterval time.Duration
}

type PollerWorker struct {
	Cfg PollerConfig
}

func (pw *PollerWorker) Run(ctx context.Context, mtr *metrics.Metrics) {
	pollerContext, cancel := context.WithCancel(ctx)
	defer cancel()

	pollTicker := time.NewTicker(pw.Cfg.PollInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-pollerContext.Done():
			return
		case <-pollTicker.C:
			mtr.UpdateMemStatsMetrics()
		}
	}
}
