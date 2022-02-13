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

type Metrics struct {
	gaugeMetrics   map[string]Gauge
	counterMetrics map[string]Counter
}

func NewMetrics() *Metrics {
	var m Metrics

	m.gaugeMetrics = make(map[string]Gauge)
	m.counterMetrics = make(map[string]Counter)

	return &m
}

func (m *Metrics) UpdateGaugeMetric(metricName string, metricData Gauge) {
	m.gaugeMetrics[metricName] = metricData
}

func (m *Metrics) UpdateCounterMetric(metricName string, metricData Counter) {
	m.counterMetrics[metricName] += metricData
}

func (m *Metrics) GetGaugeMetric(metricName string) (Gauge, bool) {
	metric, ok := m.gaugeMetrics[metricName]

	return metric, ok
}

func (m *Metrics) GetCounterMetric(metricName string) (Counter, bool) {
	metric, ok := m.counterMetrics[metricName]

	return metric, ok
}

func (m *Metrics) GetGaugeMetrics() map[string]Gauge {
	return m.gaugeMetrics
}

func (m *Metrics) GetCounterMetrics() map[string]Counter {
	return m.counterMetrics
}
