package agent

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/itd27m01/go-metrics-service/internal/models/metrics"
	"github.com/itd27m01/go-metrics-service/internal/repository"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

const (
	counterIncrement = 1
	pollTimeout      = 1 * time.Second
	sampleTime       = 1 * time.Second
)

// PollerConfig is a config for poller worker
type PollerConfig struct {
	PollInterval time.Duration `env:"POLL_INTERVAL"`
}

// PollerWorker defines poller worker type
type PollerWorker struct {
	Cfg PollerConfig
}

// RunMemStats runs worker for collecting memory metrics from localhost
func (pw *PollerWorker) RunMemStats(ctx context.Context, mtr repository.Store) {
	pollTicker := time.NewTicker(pw.Cfg.PollInterval)
	defer pollTicker.Stop()

	storeContext, storeCancel := context.WithTimeout(ctx, pollTimeout)
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

// RunPsStats runs worker for collecting processes and cpu metrics from localhost
func (pw *PollerWorker) RunPsStats(ctx context.Context, mtr repository.Store) {
	pollTicker := time.NewTicker(pw.Cfg.PollInterval)
	defer pollTicker.Stop()

	storeContext, storeCancel := context.WithCancel(ctx)
	defer storeCancel()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			UpdatePsMetrics(storeContext, mtr)
		}
	}
}

// UpdateMemStatsMetrics updates the memory metrics
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

// UpdatePsMetrics collects process and cpu metrics from localhost
func UpdatePsMetrics(ctx context.Context, mtr repository.Store) {
	updateMemPsMetrics(ctx, mtr)
	updateCPUPsMetrics(ctx, mtr)
}

// updateMemPsMetrics updates virtual memory metrics
func updateMemPsMetrics(ctx context.Context, mtr repository.Store) {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		log.Error().Msgf("Couldn't get virtual memory stats: %v", err)

		return
	}

	_ = mtr.UpdateGaugeMetric(ctx, "TotalMemory", metrics.Gauge(vm.Total))
	_ = mtr.UpdateGaugeMetric(ctx, "FreeMemory", metrics.Gauge(vm.Free))
}

// updateCPUPsMetrics updates cpu metrics
func updateCPUPsMetrics(ctx context.Context, mtr repository.Store) {
	cpuUtilization, err := cpu.PercentWithContext(ctx, sampleTime, true)
	if err != nil {
		log.Error().Msgf("Couldn't get cpu utilization: %v", err)

		return
	}

	for i, v := range cpuUtilization {
		_ = mtr.UpdateGaugeMetric(ctx, fmt.Sprintf("CPUutilization%d", i+1), metrics.Gauge(v))
	}
}
