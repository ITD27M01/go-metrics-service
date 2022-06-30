package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/itd27m01/go-metrics-service/internal/models/metrics"
)

var (
	_ Store = (*InMemoryStore)(nil)
)

// InMemoryStore implements Store interface to store metrics in memory
type InMemoryStore struct {
	metricsCache map[string]*metrics.Metric
	lock         sync.RWMutex
}

// NewInMemoryStore creates in memory store
func NewInMemoryStore() *InMemoryStore {
	var m InMemoryStore

	m.metricsCache = make(map[string]*metrics.Metric)

	return &m
}

// UpdateCounterMetric updates counter metric type
func (m *InMemoryStore) UpdateCounterMetric(_ context.Context, metricName string, metricData metrics.Counter) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	currentMetric, ok := m.metricsCache[metricName]
	switch {
	case ok && currentMetric.Delta != nil:
		*(currentMetric.Delta) += metricData
	case ok && currentMetric.Delta == nil:
		return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, currentMetric.MType)
	default:
		m.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.MetricTypeCounter,
			Delta: &metricData,
		}
	}

	return nil
}

// ResetCounterMetric resets counter to default zero value
func (m *InMemoryStore) ResetCounterMetric(_ context.Context, metricName string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

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
			MType: metrics.MetricTypeCounter,
			Delta: &zero,
		}
	}

	return nil
}

// UpdateGaugeMetric updates gauge type metric
func (m *InMemoryStore) UpdateGaugeMetric(_ context.Context, metricName string, metricData metrics.Gauge) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	currentMetric, ok := m.metricsCache[metricName]
	switch {
	case ok && currentMetric.Value != nil:
		*(currentMetric.Value) = metricData
	case ok && currentMetric.Value == nil:
		return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, currentMetric.MType)
	default:
		m.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.MetricTypeGauge,
			Value: &metricData,
		}
	}

	return nil
}

// UpdateMetrics update number of metrics
func (m *InMemoryStore) UpdateMetrics(_ context.Context, metricsBatch []*metrics.Metric) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, metric := range metricsBatch {
		currentMetric, ok := m.metricsCache[metric.ID]
		switch {
		case ok && metric.MType == metrics.MetricTypeGauge && currentMetric.Value != nil:
			currentMetric.Value = metric.Value
		case ok && metric.MType == metrics.MetricTypeGauge && currentMetric.Value == nil:
			return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metric.ID, currentMetric.MType)
		case ok && metric.MType == metrics.MetricTypeCounter && currentMetric.Delta != nil:
			*(currentMetric.Delta) += *(metric.Delta)
		case ok && metric.MType == metrics.MetricTypeCounter && currentMetric.Delta == nil:
			return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metric.ID, currentMetric.MType)
		default:
			m.metricsCache[metric.ID] = metric
		}
	}

	return nil
}

// GetMetric return metric by name
func (m *InMemoryStore) GetMetric(_ context.Context, metricName string, _ string) (*metrics.Metric, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	metric, ok := m.metricsCache[metricName]
	if !ok {
		return nil, ErrMetricNotFound
	}

	return metric, nil
}

// GetMetrics returns all of stored metrics
func (m *InMemoryStore) GetMetrics(_ context.Context) (map[string]*metrics.Metric, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	metricsData := make(map[string]*metrics.Metric)

	for k, v := range m.metricsCache {
		metricsData[k] = v
	}

	return metricsData, nil
}

// Ping checks that underlying store is alive
func (m *InMemoryStore) Ping(_ context.Context) error { return nil }
