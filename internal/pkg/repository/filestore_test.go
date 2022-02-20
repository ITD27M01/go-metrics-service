package repository

import (
	"os"
	"testing"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

const (
	testMetrics     = `{"Alloc":{"id":"Alloc","type":"gauge","value":1336312}}`
	testMetricValue = 1336312
)

func TestFileStore_LoadMetrics(t *testing.T) {
	f, _ := os.CreateTemp("", "tests")
	defer f.Close()
	defer os.Remove(f.Name())

	testMetrics := []byte(testMetrics)
	testMetricName := "Alloc"
	testMetricValue := metrics.Gauge(testMetricValue)
	f.Write(testMetrics)
	f.Seek(0, 0)

	metricsCache := make(map[string]*metrics.Metric)
	type fields struct {
		file         *os.File
		metricsCache map[string]*metrics.Metric
	}
	tests := []struct {
		name   string
		fields fields
		want   metrics.Gauge
	}{
		{
			name: testMetricName,
			fields: fields{
				file:         f,
				metricsCache: metricsCache,
			},
			want: testMetricValue,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FileStore{
				file:         tt.fields.file,
				metricsCache: tt.fields.metricsCache,
			}

			if err := fs.LoadMetrics(); err != nil || tt.want != *(fs.metricsCache[tt.name].Value) {
				t.Errorf("LoadMetrics() failed (error = %v), want %f, got %f", err, tt.want, *(fs.metricsCache[tt.name].Value))
			}
		})
	}
}

func TestFileStore_SaveMetrics(t *testing.T) {
	f, _ := os.CreateTemp("", "tests")
	defer f.Close()
	defer os.Remove(f.Name())

	testMetricName := "Alloc"
	testMetricValue := metrics.Gauge(testMetricValue)

	metricsCache := make(map[string]*metrics.Metric)
	metric := metrics.Metric{
		ID:    testMetricName,
		MType: metrics.GaugeMetricTypeName,
		Value: &testMetricValue,
	}
	metricsCache[testMetricName] = &metric
	type fields struct {
		file         *os.File
		metricsCache map[string]*metrics.Metric
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: testMetricName,
			fields: fields{
				file:         f,
				metricsCache: metricsCache,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FileStore{
				file:         tt.fields.file,
				metricsCache: tt.fields.metricsCache,
			}
			fs.SaveMetrics()

			buf := make([]byte, len(testMetrics))
			if _, err := f.ReadAt(buf, 0); err != nil || string(buf) != testMetrics {
				t.Errorf("SaveMetrics() failed (error = %v), want %s, got %v", err, testMetrics, string(buf))
			}
		})
	}
}
