package repository

import (
	"context"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

type Store interface {
	UpdateCounterMetric(name string, value metrics.Counter)
	ResetCounterMetric(name string)

	UpdateGaugeMetric(name string, value metrics.Gauge)

	GetMetric(name string) (*metrics.Metric, bool)

	GetMetrics() map[string]*metrics.Metric

	RunPreserver(ctx context.Context, syncInterval time.Duration)
	LoadMetrics() error
	Close() error
}
