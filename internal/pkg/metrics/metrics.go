package metrics

const (
	GaugeMetricTypeName   = "gauge"
	CounterMetricTypeName = "counter"
)

type Gauge float64
type Counter int64

type Metrics struct {
	GaugeMetrics   map[string]Gauge
	CounterMetrics map[string]Counter
}

func NewMetrics() *Metrics {
	var m Metrics

	m.GaugeMetrics = make(map[string]Gauge)
	m.CounterMetrics = make(map[string]Counter)

	return &m
}
