package metrics

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

const (
	MetricTypeGauge   = "gauge"
	MetricTypeCounter = "counter"
)

type Gauge float64
type Counter int64

type Metric struct {
	ID    string   `json:"id"`              // Имя метрики
	MType string   `json:"type"`            // Параметр, принимающий значение gauge или counter
	Delta *Counter `json:"delta,omitempty"` // Значение метрики в случае передачи counter
	Value *Gauge   `json:"value,omitempty"` // Значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // Значение хеш-функции
}

func (m *Metric) EncodeMetric() (*bytes.Buffer, error) {
	var buf bytes.Buffer
	jsonEncoder := json.NewEncoder(&buf)

	if err := jsonEncoder.Encode(m); err != nil {
		return nil, err
	}

	return &buf, nil
}

func (m *Metric) SetHash(key string) {
	if key == "" {
		return
	}

	m.Hash = m.getHash(key)
}

func (m *Metric) IsHashValid(key string) bool {
	if key == "" {
		return true
	}

	return m.Hash == m.getHash(key)
}

func (m *Metric) getHash(key string) string {
	var metricString string
	switch m.MType {
	case MetricTypeGauge:
		metricString = fmt.Sprintf("%s:%s:%f", m.ID, MetricTypeGauge, *(m.Value))
	case MetricTypeCounter:
		metricString = fmt.Sprintf("%s:%s:%d", m.ID, MetricTypeCounter, *(m.Delta))
	default:
		log.Error().Msgf("unsupported metric type: %s", m.MType)
	}

	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(metricString))

	return hex.EncodeToString(mac.Sum(nil))
}

func (m *Metric) String() string {
	switch m.MType {
	case MetricTypeGauge:
		return fmt.Sprintf("%g", *(m.Value))
	case MetricTypeCounter:
		return fmt.Sprintf("%d", *(m.Delta))
	default:
		return ""
	}
}
