package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

const (
	gaugeBitSize   = 64
	counterBase    = 10
	counterBitSize = 64
)

func RegisterHandlers(mux *chi.Mux, metricsServer *MetricsServer) {
	mux.Route("/update/{metricType}/{metricName}/{metricData}", UpdateHandler(metricsServer))
	mux.Route("/value/{metricType}/{metricName}", GetMetricHandler(metricsServer))
	mux.Route("/", GetMetricsHandler(metricsServer))
}

func UpdateHandler(metricsServer *MetricsServer) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			metricType := chi.URLParam(r, "metricType")
			metricName := chi.URLParam(r, "metricName")
			metricData := chi.URLParam(r, "metricData")

			var err error
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
		})
	}
}

func GetMetricHandler(metricsServer *MetricsServer) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			metricType := chi.URLParam(r, "metricType")
			metricName := chi.URLParam(r, "metricName")

			var ok bool
			var stringifyMetricData string
			switch {
			case metricType == metrics.GaugeMetricTypeName:
				var metricData metrics.Gauge
				metricData, ok = metricsServer.Cfg.MetricsData.GaugeMetrics[metricName]
				stringifyMetricData = fmt.Sprintf("%g", metricData)
			case metricType == metrics.CounterMetricTypeName:
				var metricData metrics.Counter
				metricData, ok = metricsServer.Cfg.MetricsData.CounterMetrics[metricName]
				stringifyMetricData = fmt.Sprintf("%d", metricData)
			default:
				http.Error(
					w,
					fmt.Sprintf("Metric type not implemented: %s", metricType),
					http.StatusNotImplemented,
				)
				return
			}
			if ok {
				_, err := w.Write([]byte(stringifyMetricData))
				if err != nil {
					http.Error(
						w,
						fmt.Sprintf("Something went wrong during metric get: %s", metricName),
						http.StatusInternalServerError,
					)
				}
			} else {
				http.Error(
					w,
					fmt.Sprintf("Metric not found: %s", metricName),
					http.StatusNotFound,
				)
			}
		})
	}
}

func GetMetricsHandler(metricsServer *MetricsServer) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			tmpl, err := template.New("index.html").Parse(metricsTemplate)
			if err != nil {
				http.Error(
					w,
					"Something went wrong during page template rendering",
					http.StatusInternalServerError,
				)
			}
			err = tmpl.Execute(w, metricsServer.Cfg.MetricsData)
			if err != nil {
				http.Error(
					w,
					"Something went wrong during metrics get",
					http.StatusInternalServerError,
				)
			}
		})
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
