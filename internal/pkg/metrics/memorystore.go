package metrics

type InMemoryStore struct {
	gaugeMetrics   map[string]Gauge
	counterMetrics map[string]Counter
}

func NewInMemoryStore() *InMemoryStore {
	var m InMemoryStore

	m.gaugeMetrics = make(map[string]Gauge)
	m.counterMetrics = make(map[string]Counter)

	return &m
}

func (m *InMemoryStore) UpdateGaugeMetric(metricName string, metricData Gauge) {
	m.gaugeMetrics[metricName] = metricData
}

func (m *InMemoryStore) UpdateCounterMetric(metricName string, metricData Counter) {
	m.counterMetrics[metricName] += metricData
}

func (m *InMemoryStore) GetGaugeMetric(metricName string) (Gauge, bool) {
	metric, ok := m.gaugeMetrics[metricName]

	return metric, ok
}

func (m *InMemoryStore) GetCounterMetric(metricName string) (Counter, bool) {
	metric, ok := m.counterMetrics[metricName]

	return metric, ok
}

func (m *InMemoryStore) GetGaugeMetrics() map[string]Gauge {
	return m.gaugeMetrics
}

func (m *InMemoryStore) GetCounterMetrics() map[string]Counter {
	return m.counterMetrics
}
