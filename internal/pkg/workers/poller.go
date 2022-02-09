package workers

import (
	"context"
	"math/rand"
	"runtime"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
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
			UpdateMemStatsMetrics(mtr)
		}
	}
}

func UpdateMemStatsMetrics(mtr *metrics.Metrics) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	mtr.CounterMetrics["PollCount"]++

	mtr.GaugeMetrics["Alloc"] = metrics.Gauge(memStats.Alloc)
	mtr.GaugeMetrics["BuckHashSys"] = metrics.Gauge(memStats.BuckHashSys)
	mtr.GaugeMetrics["Frees"] = metrics.Gauge(memStats.Frees)
	mtr.GaugeMetrics["GCCPUFraction"] = metrics.Gauge(memStats.GCCPUFraction)
	mtr.GaugeMetrics["GCSys"] = metrics.Gauge(memStats.GCSys)
	mtr.GaugeMetrics["HeapAlloc"] = metrics.Gauge(memStats.HeapAlloc)
	mtr.GaugeMetrics["HeapIdle"] = metrics.Gauge(memStats.HeapIdle)
	mtr.GaugeMetrics["HeapInuse"] = metrics.Gauge(memStats.HeapInuse)
	mtr.GaugeMetrics["HeapObjects"] = metrics.Gauge(memStats.HeapObjects)
	mtr.GaugeMetrics["HeapReleased"] = metrics.Gauge(memStats.HeapReleased)
	mtr.GaugeMetrics["HeapSys"] = metrics.Gauge(memStats.HeapSys)
	mtr.GaugeMetrics["LastGC"] = metrics.Gauge(memStats.LastGC)
	mtr.GaugeMetrics["Lookups"] = metrics.Gauge(memStats.Lookups)
	mtr.GaugeMetrics["MCacheInuse"] = metrics.Gauge(memStats.MCacheInuse)
	mtr.GaugeMetrics["MCacheSys"] = metrics.Gauge(memStats.MCacheSys)
	mtr.GaugeMetrics["MSpanInuse"] = metrics.Gauge(memStats.MSpanInuse)
	mtr.GaugeMetrics["MSpanSys"] = metrics.Gauge(memStats.MSpanSys)
	mtr.GaugeMetrics["Mallocs"] = metrics.Gauge(memStats.Mallocs)
	mtr.GaugeMetrics["NextGC"] = metrics.Gauge(memStats.NextGC)
	mtr.GaugeMetrics["NumForcedGC"] = metrics.Gauge(memStats.NumForcedGC)
	mtr.GaugeMetrics["NumGC"] = metrics.Gauge(memStats.NumGC)
	mtr.GaugeMetrics["OtherSys"] = metrics.Gauge(memStats.OtherSys)
	mtr.GaugeMetrics["PauseTotalNs"] = metrics.Gauge(memStats.PauseTotalNs)
	mtr.GaugeMetrics["StackInuse"] = metrics.Gauge(memStats.StackInuse)
	mtr.GaugeMetrics["StackSys"] = metrics.Gauge(memStats.StackSys)
	mtr.GaugeMetrics["Sys"] = metrics.Gauge(memStats.Sys)

	mtr.GaugeMetrics["RandomValue"] = metrics.Gauge(rand.Int63())
}
