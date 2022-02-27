package repository

import (
	"context"
	"sync"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

type InMemoryStore struct {
	metricsCache map[string]*metrics.Metric
	mu           sync.Mutex
}

func NewInMemoryStore() *InMemoryStore {
	var m InMemoryStore

	m.metricsCache = make(map[string]*metrics.Metric)

	return &m
}

func (m *InMemoryStore) UpdateCounterMetric(metricName string, metricData metrics.Counter) {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentMetric, ok := m.metricsCache[metricName]
	if ok && currentMetric.Delta != nil {
		*(currentMetric.Delta) += metricData
	} else {
		m.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.CounterMetricTypeName,
			Delta: &metricData,
		}
	}
}

func (m *InMemoryStore) ResetCounterMetric(metricName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var zero metrics.Counter
	currentMetric, ok := m.metricsCache[metricName]
	if ok {
		*(currentMetric.Delta) = zero
	} else {
		m.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.CounterMetricTypeName,
			Delta: &zero,
		}
	}
}

func (m *InMemoryStore) UpdateGaugeMetric(metricName string, metricData metrics.Gauge) {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentMetric, ok := m.metricsCache[metricName]
	if ok && currentMetric.Value != nil {
		*(currentMetric.Value) = metricData
	} else {
		m.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.GaugeMetricTypeName,
			Value: &metricData,
		}
	}
}

func (m *InMemoryStore) GetMetric(metricName string) (*metrics.Metric, bool) {
	metric, ok := m.metricsCache[metricName]

	return metric, ok
}

func (m *InMemoryStore) GetMetrics() map[string]*metrics.Metric {
	return m.metricsCache
}

func (m *InMemoryStore) LoadMetrics() error                              { return nil }
func (m *InMemoryStore) RunPreserver(_ context.Context, _ time.Duration) {}
func (m *InMemoryStore) Close() error                                    { return nil }
