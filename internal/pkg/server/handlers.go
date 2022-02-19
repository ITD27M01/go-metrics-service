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
	"github.com/itd27m01/go-metrics-service/internal/pkg/repository"
)

//go:embed assets/index.gohtml
var metricsTemplateFile string

const (
	gaugeBitSize   = 64
	counterBase    = 10
	counterBitSize = 64
)

func RegisterHandlers(mux *chi.Mux, metricsStore repository.Store) {
	mux.Route("/update/", UpdateHandler(metricsStore))
	mux.Route("/value/", GetMetricHandler(metricsStore))
	mux.Route("/", GetMetricsHandler(metricsStore))
}

func UpdateHandler(metricsStore repository.Store) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", updateHandlerJSON(metricsStore))
		r.Post("/{metricType}/{metricName}/{metricData}", updateHandlerPlain(metricsStore))
	}
}

func GetMetricHandler(metricsStore repository.Store) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", retrieveHandlerJSON(metricsStore))
		r.Get("/{metricType}/{metricName}", getHandlerPlain(metricsStore))
	}
}

func GetMetricsHandler(metricsStore repository.Store) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			metricsData := metricsStore.GetMetrics()

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

func updateHandlerJSON(metricsStore repository.Store) func(w http.ResponseWriter, r *http.Request) {
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

func updateHandlerPlain(metricsStore repository.Store) func(w http.ResponseWriter, r *http.Request) {
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

func retrieveHandlerJSON(metricsStore repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric metrics.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		switch {
		case metric.MType == metrics.GaugeMetricTypeName:
			metricData, ok := metricsStore.GetMetric(metric.ID)
			if ok && metricData.Value != nil {
				metric.Value = metricData.Value
			} else {
				http.Error(
					w,
					fmt.Sprintf("Metric not found: %s", metric.ID),
					http.StatusNotFound,
				)

				return
			}
		case metric.MType == metrics.CounterMetricTypeName:
			metricData, ok := metricsStore.GetMetric(metric.ID)
			if ok && metricData.Delta != nil {
				metric.Delta = metricData.Delta
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

func getHandlerPlain(metricsStore repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		var stringifyMetricData string
		switch {
		case metricType == metrics.GaugeMetricTypeName:
			metricData, ok := metricsStore.GetMetric(metricName)
			if ok && metricData.Value != nil {
				stringifyMetricData = fmt.Sprintf("%g", *metricData.Value)
			} else {
				http.Error(
					w,
					fmt.Sprintf("Metric not found: %s", metricName),
					http.StatusNotFound,
				)

				return
			}

		case metricType == metrics.CounterMetricTypeName:
			metricData, ok := metricsStore.GetMetric(metricName)
			if ok && metricData.Delta != nil {
				stringifyMetricData = fmt.Sprintf("%d", *metricData.Delta)
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

func updateGaugeMetric(metricName string, metricData string, metricsStore repository.Store) error {
	if parsedData, err := strconv.ParseFloat(metricData, gaugeBitSize); err == nil {
		metricsStore.UpdateGaugeMetric(metricName, metrics.Gauge(parsedData))
	} else {
		return err
	}

	return nil
}

func updateCounterMetric(metricName string, metricData string, metricsStore repository.Store) error {
	if parsedData, err := strconv.ParseInt(metricData, counterBase, counterBitSize); err == nil {
		metricsStore.UpdateCounterMetric(metricName, metrics.Counter(parsedData))
	} else {
		return err
	}

	return nil
}
