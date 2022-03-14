package agent

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

	_ = mtr.UpdateCounterMetric(ctx, "PollCount", counterIncrement)

	_ = mtr.UpdateGaugeMetric(ctx, "Alloc", metrics.Gauge(memStats.Alloc))
	_ = mtr.UpdateGaugeMetric(ctx, "BuckHashSys", metrics.Gauge(memStats.BuckHashSys))

	_ = mtr.UpdateGaugeMetric(ctx, "BuckHashSys", metrics.Gauge(memStats.BuckHashSys))
	_ = mtr.UpdateGaugeMetric(ctx, "Frees", metrics.Gauge(memStats.Frees))
	_ = mtr.UpdateGaugeMetric(ctx, "GCCPUFraction", metrics.Gauge(memStats.GCCPUFraction))
	_ = mtr.UpdateGaugeMetric(ctx, "GCSys", metrics.Gauge(memStats.GCSys))
	_ = mtr.UpdateGaugeMetric(ctx, "HeapAlloc", metrics.Gauge(memStats.HeapAlloc))
	_ = mtr.UpdateGaugeMetric(ctx, "HeapIdle", metrics.Gauge(memStats.HeapIdle))
	_ = mtr.UpdateGaugeMetric(ctx, "HeapInuse", metrics.Gauge(memStats.HeapInuse))
	_ = mtr.UpdateGaugeMetric(ctx, "HeapObjects", metrics.Gauge(memStats.HeapObjects))
	_ = mtr.UpdateGaugeMetric(ctx, "HeapReleased", metrics.Gauge(memStats.HeapReleased))
	_ = mtr.UpdateGaugeMetric(ctx, "HeapSys", metrics.Gauge(memStats.HeapSys))
	_ = mtr.UpdateGaugeMetric(ctx, "LastGC", metrics.Gauge(memStats.LastGC))
	_ = mtr.UpdateGaugeMetric(ctx, "Lookups", metrics.Gauge(memStats.Lookups))
	_ = mtr.UpdateGaugeMetric(ctx, "MCacheInuse", metrics.Gauge(memStats.MCacheInuse))
	_ = mtr.UpdateGaugeMetric(ctx, "MCacheSys", metrics.Gauge(memStats.MCacheSys))
	_ = mtr.UpdateGaugeMetric(ctx, "MSpanInuse", metrics.Gauge(memStats.MSpanInuse))
	_ = mtr.UpdateGaugeMetric(ctx, "MSpanSys", metrics.Gauge(memStats.MSpanSys))
	_ = mtr.UpdateGaugeMetric(ctx, "Mallocs", metrics.Gauge(memStats.Mallocs))
	_ = mtr.UpdateGaugeMetric(ctx, "NextGC", metrics.Gauge(memStats.NextGC))
	_ = mtr.UpdateGaugeMetric(ctx, "NumForcedGC", metrics.Gauge(memStats.NumForcedGC))
	_ = mtr.UpdateGaugeMetric(ctx, "NumGC", metrics.Gauge(memStats.NumGC))
	_ = mtr.UpdateGaugeMetric(ctx, "OtherSys", metrics.Gauge(memStats.OtherSys))
	_ = mtr.UpdateGaugeMetric(ctx, "PauseTotalNs", metrics.Gauge(memStats.PauseTotalNs))
	_ = mtr.UpdateGaugeMetric(ctx, "StackInuse", metrics.Gauge(memStats.StackInuse))
	_ = mtr.UpdateGaugeMetric(ctx, "StackSys", metrics.Gauge(memStats.StackSys))
	_ = mtr.UpdateGaugeMetric(ctx, "Sys", metrics.Gauge(memStats.Sys))
	_ = mtr.UpdateGaugeMetric(ctx, "TotalAlloc", metrics.Gauge(memStats.TotalAlloc))

	_ = mtr.UpdateGaugeMetric(ctx, "RandomValue", metrics.Gauge(rand.Int63()))
}
