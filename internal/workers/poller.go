package workers

import (
	"context"
	"math/rand"
	"runtime"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	"github.com/itd27m01/go-metrics-service/internal/repository"
)

const (
	counterIncrement = 1
	storeTimeout     = 1 * time.Second
)

type PollerConfig struct {
	PollInterval time.Duration `env:"POLL_INTERVAL"`
}

type PollerWorker struct {
	Cfg PollerConfig
}

func (pw *PollerWorker) Run(ctx context.Context, mtr repository.Store) {
	pollTicker := time.NewTicker(pw.Cfg.PollInterval)
	defer pollTicker.Stop()

	storeContext, storeCancel := context.WithTimeout(ctx, storeTimeout)
	defer storeCancel()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			UpdateMemStatsMetrics(storeContext, mtr)
		}
	}
}

func UpdateMemStatsMetrics(ctx context.Context, mtr repository.Store) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	mtr.UpdateCounterMetric(ctx, "PollCount", counterIncrement)

	mtr.UpdateGaugeMetric(ctx, "Alloc", metrics.Gauge(memStats.Alloc))
	mtr.UpdateGaugeMetric(ctx, "BuckHashSys", metrics.Gauge(memStats.BuckHashSys))

	mtr.UpdateGaugeMetric(ctx, "BuckHashSys", metrics.Gauge(memStats.BuckHashSys))
	mtr.UpdateGaugeMetric(ctx, "Frees", metrics.Gauge(memStats.Frees))
	mtr.UpdateGaugeMetric(ctx, "GCCPUFraction", metrics.Gauge(memStats.GCCPUFraction))
	mtr.UpdateGaugeMetric(ctx, "GCSys", metrics.Gauge(memStats.GCSys))
	mtr.UpdateGaugeMetric(ctx, "HeapAlloc", metrics.Gauge(memStats.HeapAlloc))
	mtr.UpdateGaugeMetric(ctx, "HeapIdle", metrics.Gauge(memStats.HeapIdle))
	mtr.UpdateGaugeMetric(ctx, "HeapInuse", metrics.Gauge(memStats.HeapInuse))
	mtr.UpdateGaugeMetric(ctx, "HeapObjects", metrics.Gauge(memStats.HeapObjects))
	mtr.UpdateGaugeMetric(ctx, "HeapReleased", metrics.Gauge(memStats.HeapReleased))
	mtr.UpdateGaugeMetric(ctx, "HeapSys", metrics.Gauge(memStats.HeapSys))
	mtr.UpdateGaugeMetric(ctx, "LastGC", metrics.Gauge(memStats.LastGC))
	mtr.UpdateGaugeMetric(ctx, "Lookups", metrics.Gauge(memStats.Lookups))
	mtr.UpdateGaugeMetric(ctx, "MCacheInuse", metrics.Gauge(memStats.MCacheInuse))
	mtr.UpdateGaugeMetric(ctx, "MCacheSys", metrics.Gauge(memStats.MCacheSys))
	mtr.UpdateGaugeMetric(ctx, "MSpanInuse", metrics.Gauge(memStats.MSpanInuse))
	mtr.UpdateGaugeMetric(ctx, "MSpanSys", metrics.Gauge(memStats.MSpanSys))
	mtr.UpdateGaugeMetric(ctx, "Mallocs", metrics.Gauge(memStats.Mallocs))
	mtr.UpdateGaugeMetric(ctx, "NextGC", metrics.Gauge(memStats.NextGC))
	mtr.UpdateGaugeMetric(ctx, "NumForcedGC", metrics.Gauge(memStats.NumForcedGC))
	mtr.UpdateGaugeMetric(ctx, "NumGC", metrics.Gauge(memStats.NumGC))
	mtr.UpdateGaugeMetric(ctx, "OtherSys", metrics.Gauge(memStats.OtherSys))
	mtr.UpdateGaugeMetric(ctx, "PauseTotalNs", metrics.Gauge(memStats.PauseTotalNs))
	mtr.UpdateGaugeMetric(ctx, "StackInuse", metrics.Gauge(memStats.StackInuse))
	mtr.UpdateGaugeMetric(ctx, "StackSys", metrics.Gauge(memStats.StackSys))
	mtr.UpdateGaugeMetric(ctx, "Sys", metrics.Gauge(memStats.Sys))
	mtr.UpdateGaugeMetric(ctx, "TotalAlloc", metrics.Gauge(memStats.TotalAlloc))

	mtr.UpdateGaugeMetric(ctx, "RandomValue", metrics.Gauge(rand.Int63()))
}
