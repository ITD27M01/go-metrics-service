package repository

import (
	"context"
	"errors"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

var ErrMetricTypeMismatch = errors.New("possible metric type mismatch")

type Store interface {
	UpdateCounterMetric(ctx context.Context, name string, value metrics.Counter) error
	ResetCounterMetric(ctx context.Context, name string) error
	UpdateGaugeMetric(ctx context.Context, name string, value metrics.Gauge) error

	GetMetric(ctx context.Context, name string, metricType string) (*metrics.Metric, bool, error)
	GetMetrics(ctx context.Context) (map[string]*metrics.Metric, error)

	Ping(ctx context.Context) error
}
