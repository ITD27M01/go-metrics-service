package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/itd27m01/go-metrics-service/internal/models/metrics"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryStore_UpdateCounterMetric(t *testing.T) {
	metricsCache := make(map[string]*metrics.Metric)
	testMetricName := "Alloc"
	testMetricValue := metrics.Counter(testMetricValue)

	type fields struct {
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
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "TestUpdateCounterMetric",
			fields: fields{
				metricsCache: metricsCache,
			},
			args: args{
				in0:        context.Background(),
				metricName: testMetricName,
				metricData: testMetricValue,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &InMemoryStore{
				metricsCache: tt.fields.metricsCache,
			}
			tt.wantErr(t, m.UpdateCounterMetric(tt.args.in0, tt.args.metricName, tt.args.metricData), fmt.Sprintf("UpdateCounterMetric(%v, %v, %v)", tt.args.in0, tt.args.metricName, tt.args.metricData))
		})
	}
}

func TestInMemoryStore_ResetCounterMetric(t *testing.T) {
	metricsCache := make(map[string]*metrics.Metric)
	testMetricName := "Alloc"

	type fields struct {
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
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "TestResetCounterMetric",
			fields: fields{
				metricsCache: metricsCache,
			},
			args: args{
				in0:        context.Background(),
				metricName: testMetricName,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &InMemoryStore{
				metricsCache: tt.fields.metricsCache,
			}
			tt.wantErr(t, m.ResetCounterMetric(tt.args.in0, tt.args.metricName), fmt.Sprintf("ResetCounterMetric(%v, %v)", tt.args.in0, tt.args.metricName))
		})
	}
}

func TestInMemoryStore_UpdateGaugeMetric(t *testing.T) {
	metricsCache := make(map[string]*metrics.Metric)
	testMetricName := "Alloc"
	testMetricValue := metrics.Gauge(testMetricValue)

	type fields struct {
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
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "TestUpdateGaugeMetric",
			fields: fields{
				metricsCache: metricsCache,
			},
			args: args{
				in0:        context.Background(),
				metricName: testMetricName,
				metricData: testMetricValue,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &InMemoryStore{
				metricsCache: tt.fields.metricsCache,
			}
			tt.wantErr(t, m.UpdateGaugeMetric(tt.args.in0, tt.args.metricName, tt.args.metricData), fmt.Sprintf("UpdateGaugeMetric(%v, %v, %v)", tt.args.in0, tt.args.metricName, tt.args.metricData))
		})
	}
}

func TestInMemoryStore_UpdateMetrics(t *testing.T) {
	metricsCache := make(map[string]*metrics.Metric)
	testMetricName := "Alloc"
	testMetricValue := metrics.Gauge(testMetricValue)

	type fields struct {
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
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "TestUpdateMetrics",
			fields: fields{
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
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &InMemoryStore{
				metricsCache: tt.fields.metricsCache,
			}
			tt.wantErr(t, m.UpdateMetrics(tt.args.in0, tt.args.metricsBatch), fmt.Sprintf("UpdateMetrics(%v, %v)", tt.args.in0, tt.args.metricsBatch))
		})
	}
}
