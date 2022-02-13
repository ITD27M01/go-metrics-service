package metrics

const (
	GaugeMetricTypeName   = "gauge"
	CounterMetricTypeName = "counter"
)

type Gauge float64
type Counter int64

type Store interface {
	UpdateCounterMetric(name string, value Counter)
	UpdateGaugeMetric(name string, value Gauge)
	GetGaugeMetric(name string) (Gauge, bool)
	GetCounterMetric(name string) (Counter, bool)
	GetGaugeMetrics() map[string]Gauge
	GetCounterMetrics() map[string]Counter
}
