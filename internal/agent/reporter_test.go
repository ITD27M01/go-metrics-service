package agent_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/itd27m01/go-metrics-service/internal/agent"
	"github.com/itd27m01/go-metrics-service/internal/models/metrics"
	"github.com/itd27m01/go-metrics-service/internal/repository"
)

const (
	updatePathLength = 3
	gaugeBitSize     = 64
	counterBase      = 10
	counterBitSize   = 64
)

func TestSendReport(t *testing.T) {
	mtr := repository.NewInMemoryStore()
	agent.UpdateMemStatsMetrics(context.Background(), mtr)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		tokens := strings.FieldsFunc(req.URL.Path, func(c rune) bool {
			return c == '/'
		})
		if len(tokens) != updatePathLength {
			t.Errorf("Metric value not provided or url malformed: %s", req.URL.Path)
			http.Error(
				rw,
				fmt.Sprintf("Metric value not provided or url malformed: %s", req.URL.Path),
				http.StatusNotFound,
			)

			return
		}
		metricType := tokens[0]
		metricData := tokens[len(tokens)-1]

		switch {
		case metricType == metrics.MetricTypeGauge:
			if _, err := strconv.ParseFloat(metricData, gaugeBitSize); err != nil {
				t.Errorf("Cannot save provided data: %s", metricData)
				http.Error(rw, fmt.Sprintf("Cannot save provided data: %s", metricData), http.StatusBadRequest)
			}

		case metricType == metrics.MetricTypeCounter:
			if _, err := strconv.ParseInt(metricData, counterBase, counterBitSize); err != nil {
				t.Errorf("Cannot save provided data: %s", metricData)
				http.Error(rw, fmt.Sprintf("Cannot save provided data: %s", metricData), http.StatusBadRequest)
			}
		default:
			t.Errorf("Metric type not implemented: %s", metricType)
			http.Error(
				rw,
				fmt.Sprintf("Metric type not implemented: %s", metricType),
				http.StatusNotImplemented,
			)
		}
	}))
	defer server.Close()

	agent.SendHTTPReport(context.Background(), mtr, server.URL, server.Client())
}

func TestSendReportJSON(t *testing.T) {
	mtr := repository.NewInMemoryStore()
	agent.UpdateMemStatsMetrics(context.Background(), mtr)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var metric metrics.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		switch {
		case metric.MType == metrics.MetricTypeGauge:
			if m, err := mtr.GetMetric(context.Background(), metric.ID, ""); err != nil || *m.Value != *metric.Value {
				t.Errorf("Metric data mismatch: %f and %f", *m.Value, *metric.Value)
				http.Error(w, fmt.Sprintf("Metric data mismatch: %f and %f", *m.Value, *metric.Value), http.StatusBadRequest)
			}

		case metric.MType == metrics.MetricTypeCounter:
			if m, err := mtr.GetMetric(context.Background(), metric.ID, ""); err != nil || *m.Delta != *metric.Delta {
				t.Errorf("Metric data mismatch: %d and %d", *m.Delta, *metric.Delta)
				http.Error(w, fmt.Sprintf("Metric data mismatch: %d and %d", *m.Delta, *metric.Delta), http.StatusBadRequest)
			}
		default:
			t.Errorf("Metric type not implemented: %s", metric.MType)
			http.Error(
				w,
				fmt.Sprintf("Metric type not implemented: %s", metric.MType),
				http.StatusNotImplemented,
			)
		}
	}))
	defer server.Close()

	agent.SendHTTPReportJSON(context.Background(), mtr, server.URL, server.Client(), "")
}
