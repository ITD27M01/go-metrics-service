package repository

import (
	"context"
	"fmt"
	"sync"

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

func (m *InMemoryStore) UpdateCounterMetric(_ context.Context, metricName string, metricData metrics.Counter) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentMetric, ok := m.metricsCache[metricName]
	switch {
	case ok && currentMetric.Delta != nil:
		*(currentMetric.Delta) += metricData
	case ok && currentMetric.Delta == nil:
		return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, currentMetric.MType)
	default:
		m.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.CounterMetricTypeName,
			Delta: &metricData,
		}
	}

	return nil
}

func (m *InMemoryStore) ResetCounterMetric(_ context.Context, metricName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var zero metrics.Counter
	currentMetric, ok := m.metricsCache[metricName]
	switch {
	case ok && currentMetric.Delta != nil:
		*(currentMetric.Delta) = zero
	case ok && currentMetric.Delta == nil:
		return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, currentMetric.MType)
	default:
		m.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.CounterMetricTypeName,
			Delta: &zero,
		}
	}

	return nil
}

func (m *InMemoryStore) UpdateGaugeMetric(_ context.Context, metricName string, metricData metrics.Gauge) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentMetric, ok := m.metricsCache[metricName]
	switch {
	case ok && currentMetric.Value != nil:
		*(currentMetric.Value) = metricData
	case ok && currentMetric.Value == nil:
		return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, currentMetric.MType)
	default:
		m.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.GaugeMetricTypeName,
			Value: &metricData,
		}
	}

	return nil
}

func (m *InMemoryStore) GetMetric(_ context.Context, metricName string, _ string) (*metrics.Metric, bool, error) {
	metric, ok := m.metricsCache[metricName]

	return metric, ok, nil
}

func (m *InMemoryStore) GetMetrics(_ context.Context) (map[string]*metrics.Metric, error) {
	return m.metricsCache, nil
}

func (m *InMemoryStore) Ping(_ context.Context) error { return nil }
