package metrics

import (
	"bytes"
	"reflect"
	"testing"
)

func TestMetric_EncodeMetric(t *testing.T) {
	gaugeValue := Gauge(96969.519)
	var buf bytes.Buffer
	buf.WriteString("{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":96969.519}\n")

	type fields struct {
		ID    string
		MType string
		Delta *Counter
		Value *Gauge
		Hash  string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *bytes.Buffer
		wantErr bool
	}{
		{
			name: "TestMetricEncode",
			fields: fields{
				ID:    "Alloc",
				MType: MetricTypeGauge,
				Value: &gaugeValue,
			},
			want:    &buf,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				ID:    tt.fields.ID,
				MType: tt.fields.MType,
				Delta: tt.fields.Delta,
				Value: tt.fields.Value,
				Hash:  tt.fields.Hash,
			}
			got, err := m.EncodeMetric()
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeMetric() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetric_IsHashValid(t *testing.T) {
	gaugeValue := Gauge(96969.519)

	type fields struct {
		ID    string
		MType string
		Delta *Counter
		Value *Gauge
		Hash  string
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "TestMetricHash",
			fields: fields{
				ID:    "Alloc",
				MType: MetricTypeGauge,
				Value: &gaugeValue,
			},
			args: args{key: "test"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				ID:    tt.fields.ID,
				MType: tt.fields.MType,
				Delta: tt.fields.Delta,
				Value: tt.fields.Value,
				Hash:  tt.fields.Hash,
			}
			m.SetHash("test")
			if got := m.IsHashValid(tt.args.key); got != tt.want {
				t.Errorf("IsHashValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
