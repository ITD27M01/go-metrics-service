package server

import (
	_ "embed" // Use templates from file to render pages
	"encoding/json"
	"fmt"
	"html/template"
	"log"
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
	mux.Route("/update/", UpdateHandler(metricsStore))
	mux.Route("/value/", GetMetricHandler(metricsStore))
	mux.Route("/", GetMetricsHandler(metricsStore))
}

func UpdateHandler(metricsStore metrics.Store) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", updateHandlerJSON(metricsStore))
		r.Post("/{metricType}/{metricName}/{metricData}", updateHandlerPlain(metricsStore))
	}
}

func GetMetricHandler(metricsStore metrics.Store) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", retrieveHandlerJSON(metricsStore))
		r.Get("/{metricType}/{metricName}", getHandlerPlain(metricsStore))
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

func updateHandlerJSON(metricsStore metrics.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric metrics.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		switch {
		case metric.MType == metrics.GaugeMetricTypeName:
			metricsStore.UpdateGaugeMetric(metric.ID, *metric.Value)
		case metric.MType == metrics.CounterMetricTypeName:
			metricsStore.UpdateCounterMetric(metric.ID, *metric.Delta)
		default:
			http.Error(
				w,
				fmt.Sprintf("Metric type not implemented: %s", metric.MType),
				http.StatusNotImplemented,
			)
		}
	}
}

func updateHandlerPlain(metricsStore metrics.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricData := chi.URLParam(r, "metricData")

		var err error
		switch {
		case metricType == metrics.GaugeMetricTypeName:
			err = updateGaugeMetric(metricName, metricData, metricsStore)
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
	}
}

func retrieveHandlerJSON(metricsStore metrics.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric metrics.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		switch {
		case metric.MType == metrics.GaugeMetricTypeName:
			metricData, ok := metricsStore.GetGaugeMetric(metric.ID)
			if ok {
				metric.Value = &metricData
			} else {
				http.Error(
					w,
					fmt.Sprintf("Metric not found: %s", metric.ID),
					http.StatusNotFound,
				)

				return
			}
		case metric.MType == metrics.CounterMetricTypeName:
			metricData, ok := metricsStore.GetCounterMetric(metric.ID)
			if ok {
				metric.Delta = &metricData
			} else {
				http.Error(
					w,
					fmt.Sprintf("Metric not found: %s", metric.ID),
					http.StatusNotFound,
				)

				return
			}
		default:
			http.Error(
				w,
				fmt.Sprintf("Metric type not implemented: %s", metric.MType),
				http.StatusNotImplemented,
			)

			return
		}
		encodedMetric, err := metric.EncodeMetric()
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot encode metric data: %q", err), http.StatusBadRequest)

			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(encodedMetric.Bytes())
		if err != nil {
			log.Printf("Cannot send request: %q", err)
		}
	}
}

func getHandlerPlain(metricsStore metrics.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		var stringifyMetricData string
		switch {
		case metricType == metrics.GaugeMetricTypeName:
			var metricData metrics.Gauge
			metricData, ok := metricsStore.GetGaugeMetric(metricName)
			if ok {
				stringifyMetricData = fmt.Sprintf("%g", metricData)
			} else {
				http.Error(
					w,
					fmt.Sprintf("Metric not found: %s", metricName),
					http.StatusNotFound,
				)

				return
			}

		case metricType == metrics.CounterMetricTypeName:
			var metricData metrics.Counter
			metricData, ok := metricsStore.GetCounterMetric(metricName)
			if ok {
				stringifyMetricData = fmt.Sprintf("%d", metricData)
			} else {
				http.Error(
					w,
					fmt.Sprintf("Metric not found: %s", metricName),
					http.StatusNotFound,
				)

				return
			}
		default:
			http.Error(
				w,
				fmt.Sprintf("Metric type not implemented: %s", metricType),
				http.StatusNotImplemented,
			)

			return
		}

		_, err := w.Write([]byte(stringifyMetricData))
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Something went wrong during metric get: %s", metricName),
				http.StatusInternalServerError,
			)
		}
	}
}

func updateGaugeMetric(metricName string, metricData string, metricsStore metrics.Store) error {
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
