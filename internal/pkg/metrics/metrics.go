package metrics

import (
	"math/rand"
	"runtime"
)

type gauge float64
type counter int64

type Metrics struct {
	GaugeMetrics   map[string]gauge
	CounterMetrics map[string]counter
}

func NewMetrics() *Metrics {
	var m Metrics

	m.GaugeMetrics = make(map[string]gauge)
	m.CounterMetrics = make(map[string]counter)

	return &m
}

func (m *Metrics) UpdateMemStatsMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.CounterMetrics["PollCount"] += 1

	m.GaugeMetrics["Alloc"] = gauge(memStats.Alloc)
	m.GaugeMetrics["BuckHashSys"] = gauge(memStats.BuckHashSys)
	m.GaugeMetrics["Frees"] = gauge(memStats.Frees)
	m.GaugeMetrics["GCCPUFraction"] = gauge(memStats.GCCPUFraction)
	m.GaugeMetrics["GCSys"] = gauge(memStats.GCSys)
	m.GaugeMetrics["HeapAlloc"] = gauge(memStats.HeapAlloc)
	m.GaugeMetrics["HeapIdle"] = gauge(memStats.HeapIdle)
	m.GaugeMetrics["HeapInuse"] = gauge(memStats.HeapInuse)
	m.GaugeMetrics["HeapObjects"] = gauge(memStats.HeapObjects)
	m.GaugeMetrics["HeapReleased"] = gauge(memStats.HeapReleased)
	m.GaugeMetrics["HeapSys"] = gauge(memStats.HeapSys)
	m.GaugeMetrics["LastGC"] = gauge(memStats.LastGC)
	m.GaugeMetrics["Lookups"] = gauge(memStats.Lookups)
	m.GaugeMetrics["MCacheInuse"] = gauge(memStats.MCacheInuse)
	m.GaugeMetrics["MCacheSys"] = gauge(memStats.MCacheSys)
	m.GaugeMetrics["MSpanInuse"] = gauge(memStats.MSpanInuse)
	m.GaugeMetrics["MSpanSys"] = gauge(memStats.MSpanSys)
	m.GaugeMetrics["Mallocs"] = gauge(memStats.Mallocs)
	m.GaugeMetrics["NextGC"] = gauge(memStats.NextGC)
	m.GaugeMetrics["NumForcedGC"] = gauge(memStats.NumForcedGC)
	m.GaugeMetrics["NumGC"] = gauge(memStats.NumGC)
	m.GaugeMetrics["OtherSys"] = gauge(memStats.OtherSys)
	m.GaugeMetrics["PauseTotalNs"] = gauge(memStats.PauseTotalNs)
	m.GaugeMetrics["StackInuse"] = gauge(memStats.StackInuse)
	m.GaugeMetrics["StackSys"] = gauge(memStats.StackSys)
	m.GaugeMetrics["Sys"] = gauge(memStats.Sys)

	m.GaugeMetrics["RandomValue"] = gauge(rand.Int63())
}
