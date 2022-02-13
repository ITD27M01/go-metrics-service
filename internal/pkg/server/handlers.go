package server

import (
	_ "embed" // Use templates from file to render pages
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

//go:embed assets/index.gohtml
var metricsTemplateFile string

const (
	gaugeBitSize   = 64
	counterBase    = 10
	counterBitSize = 64
)

func RegisterHandlers(mux *chi.Mux, metricsStore metrics.Store) {
	mux.Route("/update/{metricType}/{metricName}/{metricData}", UpdateHandler(metricsStore))
	mux.Route("/value/{metricType}/{metricName}", GetMetricHandler(metricsStore))
	mux.Route("/", GetMetricsHandler(metricsStore))
}

func UpdateHandler(metricsStore metrics.Store) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			metricType := chi.URLParam(r, "metricType")
			metricName := chi.URLParam(r, "metricName")
			metricData := chi.URLParam(r, "metricData")

			var err error
			switch {
			case metricType == metrics.GaugeMetricTypeName:
				err = updateGageMetric(metricName, metricData, metricsStore)
			case metricType == metrics.CounterMetricTypeName:
				err = updateCounterMetric(metricName, metricData, metricsStore)
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

func GetMetricHandler(metricsStore metrics.Store) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			metricType := chi.URLParam(r, "metricType")
			metricName := chi.URLParam(r, "metricName")

			var ok bool
			var stringifyMetricData string
			switch {
			case metricType == metrics.GaugeMetricTypeName:
				var metricData metrics.Gauge
				metricData, ok = metricsStore.GetGaugeMetric(metricName)
				stringifyMetricData = fmt.Sprintf("%g", metricData)
			case metricType == metrics.CounterMetricTypeName:
				var metricData metrics.Counter
				metricData, ok = metricsStore.GetCounterMetric(metricName)
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

func GetMetricsHandler(metricsStore metrics.Store) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			metricsData := struct {
				Gauge   map[string]metrics.Gauge
				Counter map[string]metrics.Counter
			}{
				Gauge:   metricsStore.GetGaugeMetrics(),
				Counter: metricsStore.GetCounterMetrics(),
			}
			tmpl, err := template.New("index.html").Parse(metricsTemplateFile)
			if err != nil {
				http.Error(
					w,
					"Something went wrong during page template rendering",
					http.StatusInternalServerError,
				)
			}
			err = tmpl.Execute(w, metricsData)
			if err != nil {
				http.Error(
					w,
					fmt.Sprintf("Something went wrong during metrics get: %q", err),
					http.StatusInternalServerError,
				)
			}
		})
	}
}

func updateGageMetric(metricName string, metricData string, metricsStore metrics.Store) error {
	if parsedData, err := strconv.ParseFloat(metricData, gaugeBitSize); err == nil {
		metricsStore.UpdateGaugeMetric(metricName, metrics.Gauge(parsedData))
	} else {
		return err
	}

	return nil
}

func updateCounterMetric(metricName string, metricData string, metricsStore metrics.Store) error {
	if parsedData, err := strconv.ParseInt(metricData, counterBase, counterBitSize); err == nil {
		metricsStore.UpdateCounterMetric(metricName, metrics.Counter(parsedData))
	} else {
		return err
	}

	return nil
}
