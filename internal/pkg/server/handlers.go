package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

const (
	updatePathLength = 4
	gaugeBitSize     = 64
	counterBase      = 10
	counterBitSize   = 64
)

func registerHandlers(mux *http.ServeMux, metricsServer *MetricsServer) {
	mux.HandleFunc("/update/", UpdateHandler(metricsServer))
}

func UpdateHandler(metricsServer *MetricsServer) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(
				w,
				fmt.Sprintf("Only POST requests are allowed, got %s", req.Method),
				http.StatusMethodNotAllowed,
			)

			return
		}
		log.Println(req.URL.Path)
		tokens := strings.FieldsFunc(req.URL.Path, func(c rune) bool {
			return c == '/'
		})
		if len(tokens) != updatePathLength {
			http.Error(
				w,
				fmt.Sprintf("Metric value not provided or url malformed: %s", req.URL.Path),
				http.StatusNotFound,
			)

			return
		}

		var err error
		metricType := tokens[1]
		metricName := tokens[2]
		metricData := tokens[3]

		switch {
		case metricType == metrics.GaugeMetricTypeName:
			err = updateGageMetric(metricName, metricData, metricsServer.Cfg.MetricsData)
		case metricType == metrics.CounterMetricTypeName:
			err = updateCounterMetric(metricName, metricData, metricsServer.Cfg.MetricsData)
		default:
			http.Error(
				w,
				fmt.Sprintf("Metric type not implemented: %s", metricType),
				http.StatusNotImplemented,
			)
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot save provided data: %s", metricData), http.StatusBadRequest)
		}
	}
}

func updateGageMetric(metricName string, metricData string, metricsData *metrics.Metrics) error {
	if parsedData, err := strconv.ParseFloat(metricData, gaugeBitSize); err == nil {
		metricsData.GaugeMetrics[metricName] = metrics.Gauge(parsedData)
	} else {
		return err
	}

	return nil
}

func updateCounterMetric(metricName string, metricData string, metricsData *metrics.Metrics) error {
	if parsedData, err := strconv.ParseInt(metricData, counterBase, counterBitSize); err == nil {
		metricsData.CounterMetrics[metricName] += metrics.Counter(parsedData)
	} else {
		return err
	}

	return nil
}
