package repository

import (
	"context"
	"errors"

	"github.com/itd27m01/go-metrics-service/internal/models/metrics"
)

// Errors for store interface type
var (
	ErrMetricTypeMismatch = errors.New("possible metric type mismatch")
	ErrMetricNotFound     = errors.New("metric not found in repository")
)

// Store defines interface type for metrics store
type Store interface {
	UpdateCounterMetric(ctx context.Context, name string, value metrics.Counter) error
	ResetCounterMetric(ctx context.Context, name string) error
	UpdateGaugeMetric(ctx context.Context, name string, value metrics.Gauge) error

	UpdateMetrics(ctx context.Context, metricsBatch []*metrics.Metric) error

	GetMetric(ctx context.Context, name string, metricType string) (*metrics.Metric, error)
	GetMetrics(ctx context.Context) (map[string]*metrics.Metric, error)

	Ping(ctx context.Context) error
}
