package repository

import (
	"context"
	"os"
	"testing"

	"github.com/itd27m01/go-metrics-service/internal/models/metrics"
	"github.com/stretchr/testify/assert"
)

const (
	testMetrics      = `{"Alloc":{"id":"Alloc","type":"gauge","value":1336312}}`
	testMetrics2     = `{"Alloc":{"id":"Alloc","type":"gauge","value":1336313}}`
	testMetricValue  = 1336312
	testMetricValue2 = 1336313
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
		MType: metrics.MetricTypeGauge,
		Value: &testMetricValue,
	}
	metricsCache[testMetricName] = &metric

	testMetricValue2 := metrics.Gauge(testMetricValue2)
	metricsCache2 := make(map[string]*metrics.Metric)
	metric2 := metrics.Metric{
		ID:    testMetricName,
		MType: metrics.MetricTypeGauge,
		Value: &testMetricValue2,
	}
	metricsCache2[testMetricName] = &metric2

	type fields struct {
		file         *os.File
		metricsCache map[string]*metrics.Metric
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Empty test",
			fields: fields{
				file:         f,
				metricsCache: metricsCache2,
			},
			want: testMetrics2,
		},
		{
			name: testMetricName,
			fields: fields{
				file:         f,
				metricsCache: metricsCache,
			},
			want: testMetrics,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FileStore{
				file:         tt.fields.file,
				metricsCache: tt.fields.metricsCache,
			}
			fs.SaveMetrics()

			buf := make([]byte, len(tt.want))
			if _, err := f.ReadAt(buf, 0); err != nil || string(buf) != tt.want {
				t.Errorf("SaveMetrics() failed (error = %v), want %s, got %v", err, tt.want, string(buf))
			}
		})
	}
}

func TestFileStore_UpdateCounterMetric(t *testing.T) {
	f, _ := os.CreateTemp("", "tests")
	defer f.Close()
	defer os.Remove(f.Name())

	testMetrics := []byte(testMetrics)
	testMetricName := "Alloc"
	testMetricValue := metrics.Counter(testMetricValue)
	f.Write(testMetrics)
	f.Seek(0, 0)

	metricsCache := make(map[string]*metrics.Metric)

	type fields struct {
		file         *os.File
		syncChannel  chan struct{}
		metricsCache map[string]*metrics.Metric
	}
	type args struct {
		in0        context.Context
		metricName string
		metricData metrics.Counter
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestUpdateCounterMetric",
			fields: fields{
				file:         f,
				syncChannel:  make(chan struct{}, 1),
				metricsCache: metricsCache,
			},
			args: args{
				in0:        context.Background(),
				metricName: testMetricName,
				metricData: testMetricValue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FileStore{
				file:         tt.fields.file,
				syncChannel:  tt.fields.syncChannel,
				metricsCache: tt.fields.metricsCache,
			}
			if err := fs.UpdateCounterMetric(tt.args.in0, tt.args.metricName, tt.args.metricData); (err != nil) != tt.wantErr {
				t.Errorf("UpdateCounterMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileStore_ResetCounterMetric(t *testing.T) {
	f, _ := os.CreateTemp("", "tests")
	defer f.Close()
	defer os.Remove(f.Name())

	testMetrics := []byte(testMetrics)
	testMetricName := "Alloc"
	f.Write(testMetrics)
	f.Seek(0, 0)

	metricsCache := make(map[string]*metrics.Metric)

	type fields struct {
		file         *os.File
		syncChannel  chan struct{}
		metricsCache map[string]*metrics.Metric
	}
	type args struct {
		in0        context.Context
		metricName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestUpdateCounterMetric",
			fields: fields{
				file:         f,
				syncChannel:  make(chan struct{}, 1),
				metricsCache: metricsCache,
			},
			args: args{
				in0:        context.Background(),
				metricName: testMetricName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FileStore{
				file:         tt.fields.file,
				syncChannel:  tt.fields.syncChannel,
				metricsCache: tt.fields.metricsCache,
			}
			if err := fs.ResetCounterMetric(tt.args.in0, tt.args.metricName); (err != nil) != tt.wantErr {
				t.Errorf("ResetCounterMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileStore_UpdateGaugeMetric(t *testing.T) {
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
		syncChannel  chan struct{}
		metricsCache map[string]*metrics.Metric
	}
	type args struct {
		in0        context.Context
		metricName string
		metricData metrics.Gauge
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestUpdateGaugeMetric",
			fields: fields{
				file:         f,
				syncChannel:  make(chan struct{}, 1),
				metricsCache: metricsCache,
			},
			args: args{
				in0:        context.Background(),
				metricName: testMetricName,
				metricData: testMetricValue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FileStore{
				file:         tt.fields.file,
				syncChannel:  tt.fields.syncChannel,
				metricsCache: tt.fields.metricsCache,
			}
			if err := fs.UpdateGaugeMetric(tt.args.in0, tt.args.metricName, tt.args.metricData); (err != nil) != tt.wantErr {
				t.Errorf("UpdateGaugeMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileStore_UpdateMetrics(t *testing.T) {
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
		syncChannel  chan struct{}
		metricsCache map[string]*metrics.Metric
	}
	type args struct {
		in0          context.Context
		metricsBatch []*metrics.Metric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "TestUpdateMetrics",
			fields: fields{
				file:         f,
				syncChannel:  make(chan struct{}, 1),
				metricsCache: metricsCache,
			},
			args: args{
				in0: context.Background(),
				metricsBatch: []*metrics.Metric{
					{
						ID:    testMetricName,
						Value: &testMetricValue,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FileStore{
				file:         tt.fields.file,
				syncChannel:  tt.fields.syncChannel,
				metricsCache: tt.fields.metricsCache,
			}
			if err := fs.UpdateMetrics(tt.args.in0, tt.args.metricsBatch); (err != nil) != tt.wantErr {
				t.Errorf("UpdateMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileStore_GetMetric(t *testing.T) {
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
		syncChannel  chan struct{}
		metricsCache map[string]*metrics.Metric
	}
	type args struct {
		in0        context.Context
		metricName string
		in2        string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *metrics.Metric
		wantErr bool
	}{
		{
			name: "TestGetMetric",
			fields: fields{
				file:         f,
				syncChannel:  make(chan struct{}, 1),
				metricsCache: metricsCache,
			},
			args: args{
				in0:        context.Background(),
				metricName: testMetricName,
			},
			want: &metrics.Metric{
				ID:    testMetricName,
				Value: &testMetricValue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FileStore{
				file:         tt.fields.file,
				syncChannel:  tt.fields.syncChannel,
				metricsCache: tt.fields.metricsCache,
			}
			_ = fs.UpdateGaugeMetric(context.Background(), tt.args.metricName, testMetricValue)
			got, err := fs.GetMetric(tt.args.in0, tt.args.metricName, tt.args.in2)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, *got.Value, *tt.want.Value) {
				t.Errorf("GetMetric() got = %v, want %v", *got.Value, *tt.want.Value)
			}
		})
	}
}
