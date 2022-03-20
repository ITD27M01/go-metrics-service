package server

import (
	"context"
	"database/sql/driver"
	_ "embed" // Use templates from file to render pages
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/itd27m01/go-metrics-service/internal/pkg/logging/log"
	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	"github.com/itd27m01/go-metrics-service/internal/repository"
)

//go:embed assets/index.gohtml
var metricsTemplateFile string

const (
	requestTimeout = 1 * time.Second
	gaugeBitSize   = 64
	counterBase    = 10
	counterBitSize = 64
)

func RegisterHandlers(mux *chi.Mux, metricsStore repository.Store, signKey string) {
	mux.Route("/ping", PingHandler(metricsStore))
	mux.Route("/update/", UpdateHandler(metricsStore, signKey))
	mux.Route("/updates/", UpdatesHandler(metricsStore))
	mux.Route("/value/", GetMetricHandler(metricsStore, signKey))
	mux.Route("/", GetMetricsHandler(metricsStore))
}

func PingHandler(metricsStore driver.Pinger) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
			defer requestCancel()

			if err := metricsStore.Ping(requestContext); err != nil {
				http.Error(
					w,
					fmt.Sprintf("Something went wrong during server ping: %q", err),
					http.StatusInternalServerError,
				)
			}
		})
	}
}

func UpdateHandler(metricsStore repository.Store, signKey string) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", updateHandlerJSON(metricsStore, signKey))
		r.Post("/{metricType}/{metricName}/{metricData}", updateHandlerPlain(metricsStore))
	}
}

func UpdatesHandler(metricsStore repository.Store) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", updatesBatchHandler(metricsStore))
	}
}

func GetMetricHandler(metricsStore repository.Store, signKey string) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", retrieveHandlerJSON(metricsStore, signKey))
		r.Get("/{metricType}/{metricName}", getHandlerPlain(metricsStore))
	}
}

func GetMetricsHandler(metricsStore repository.Store) func(r chi.Router) {
	var tmpl = template.Must(template.New("index.html").Parse(metricsTemplateFile))

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
			defer requestCancel()

			metricsData, err := metricsStore.GetMetrics(requestContext)
			if err != nil {
				http.Error(
					w,
					fmt.Sprintf("Something went wrong during metrics get: %q", err),
					http.StatusInternalServerError,
				)
			}

			w.Header().Set("Content-Type", "text/html")
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

func updateHandlerJSON(metricsStore repository.Store, signKey string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
		defer requestCancel()

		var metric metrics.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		if !metric.IsHashValid(signKey) {
			http.Error(w, "Wrong hash provided for metric", http.StatusBadRequest)

			return
		}

		switch {
		case metric.MType == metrics.MetricTypeGauge:
			if metric.Value == nil {
				http.Error(
					w,
					"Value is required field",
					http.StatusBadRequest,
				)
			}
			err := metricsStore.UpdateGaugeMetric(requestContext, metric.ID, *metric.Value)
			if err != nil {
				http.Error(
					w,
					fmt.Sprintf("Failed to update metric: %q", err),
					http.StatusBadRequest,
				)
			}
		case metric.MType == metrics.MetricTypeCounter:
			if metric.Delta == nil {
				http.Error(
					w,
					"Delta is required field",
					http.StatusBadRequest,
				)
			}
			err := metricsStore.UpdateCounterMetric(requestContext, metric.ID, *(metric.Delta))
			if err != nil {
				http.Error(
					w,
					fmt.Sprintf("Failed to update metric: %q", err),
					http.StatusBadRequest,
				)
			}
		default:
			http.Error(
				w,
				fmt.Sprintf("Metric type not implemented: %s", metric.MType),
				http.StatusNotImplemented,
			)
		}
	}
}

func updatesBatchHandler(metricsStore repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
		defer requestCancel()

		var metricsSlice []*metrics.Metric
		err := json.NewDecoder(r.Body).Decode(&metricsSlice)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		err = metricsStore.UpdateMetrics(requestContext, metricsSlice)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Failed to update metrics: %q", err),
				http.StatusBadRequest,
			)
		}
	}
}

func updateHandlerPlain(metricsStore repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricData := chi.URLParam(r, "metricData")

		requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
		defer requestCancel()

		var err error
		switch {
		case metricType == metrics.MetricTypeGauge:
			err = updateGauge(requestContext, metricName, metricData, metricsStore)
		case metricType == metrics.MetricTypeCounter:
			err = updateCounter(requestContext, metricName, metricData, metricsStore)
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

func retrieveHandlerJSON(metricsStore repository.Store, signKey string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric metrics.Metric
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
		defer requestCancel()

		metricData, err := metricsStore.GetMetric(requestContext, metric.ID, metric.MType)
		switch {
		case errors.Is(err, repository.ErrMetricNotFound):
			http.Error(
				w,
				fmt.Sprintf("Metric not found: %s", metric.ID),
				http.StatusNotFound,
			)

			return
		case !errors.Is(err, nil):
			http.Error(
				w,
				fmt.Sprintf("Filed to get metric: %q", err),
				http.StatusInternalServerError,
			)

			return
		}

		metricData.SetHash(signKey)

		encodedMetric, err := metricData.EncodeMetric()
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot encode metric data: %q", err), http.StatusInternalServerError)

			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(encodedMetric.Bytes())
		if err != nil {
			log.Error().Err(err).Msg("Cannot send request")
		}
	}
}

func getHandlerPlain(metricsStore repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricType != metrics.MetricTypeGauge && metricType != metrics.MetricTypeCounter {
			http.Error(
				w,
				fmt.Sprintf("Metric type not implemented: %s", metricType),
				http.StatusNotImplemented,
			)

			return
		}

		requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
		defer requestCancel()

		metricData, err := metricsStore.GetMetric(requestContext, metricName, metricType)
		switch {
		case errors.Is(err, repository.ErrMetricNotFound):
			http.Error(
				w,
				fmt.Sprintf("Metric not found: %s", metricName),
				http.StatusNotFound,
			)

			return
		case !errors.Is(err, nil):
			http.Error(
				w,
				fmt.Sprintf("Filed to get metric: %q", err),
				http.StatusInternalServerError,
			)

			return
		}

		_, err = w.Write([]byte(metricData.String()))
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Something went wrong during metric get: %s", metricName),
				http.StatusInternalServerError,
			)
		}
	}
}

func updateGauge(ctx context.Context, metricName string, metricData string, metricsStore repository.Store) error {
	parsedData, err := strconv.ParseFloat(metricData, gaugeBitSize)
	if err == nil {
		return metricsStore.UpdateGaugeMetric(ctx, metricName, metrics.Gauge(parsedData))
	}

	return err
}

func updateCounter(ctx context.Context, metricName string, metricData string, metricsStore repository.Store) error {
	parsedData, err := strconv.ParseInt(metricData, counterBase, counterBitSize)
	if err == nil {
		return metricsStore.UpdateCounterMetric(ctx, metricName, metrics.Counter(parsedData))
	}

	return err
}
