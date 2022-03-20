package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/pkg/logging/log"
	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	"github.com/itd27m01/go-metrics-service/internal/repository"
)

type ReporterConfig struct {
	ServerScheme   string `env:"SERVER_SCHEME" envDefault:"http"`
	ServerAddress  string `env:"ADDRESS"`
	ServerPath     string `env:"SERVER_PATH" envDefault:"/update/"`
	ServerTimeout  time.Duration
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	SignKey        string        `env:"KEY"`
}

type ReportWorker struct {
	Cfg ReporterConfig
}

func (rw *ReportWorker) Run(ctx context.Context, mtr repository.Store) {
	reportTicker := time.NewTicker(rw.Cfg.ReportInterval)
	defer reportTicker.Stop()

	client := http.Client{
		Timeout: rw.Cfg.ServerTimeout,
	}

	serverURL := rw.Cfg.ServerScheme + "://" + rw.Cfg.ServerAddress
	sendURL := serverURL + rw.Cfg.ServerPath

	for {
		select {
		case <-ctx.Done():
			return
		case <-reportTicker.C:
			SendReport(ctx, mtr, sendURL, &client)
			SendReportJSON(ctx, mtr, sendURL, &client, rw.Cfg.SignKey)
			SendBatchJSON(ctx, mtr, serverURL, &client)
			resetCounters(ctx, mtr)
		}
	}
}

func SendReport(ctx context.Context, mtr repository.Store, serverURL string, client *http.Client) {
	getContext, getCancel := context.WithTimeout(ctx, storeTimeout)
	defer getCancel()

	metricsMap, err := mtr.GetMetrics(getContext)
	if err != nil {
		log.Error().Err(err).Msgf("Some error occurred during metrics get")
	}

	serverURL = strings.TrimSuffix(serverURL, "/")
	var stringifyMetricValue string

	for _, v := range metricsMap {
		if v.MType == metrics.MetricTypeGauge {
			stringifyMetricValue = fmt.Sprintf("%f", *v.Value)
		} else {
			stringifyMetricValue = fmt.Sprintf("%d", *v.Delta)
		}
		metricUpdateURL := fmt.Sprintf("%s/%s/%s/%s", serverURL, v.MType, v.ID, stringifyMetricValue)
		err := sendMetric(ctx, metricUpdateURL, client)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to send metric %s", v.ID)
		}
	}
}

func SendReportJSON(ctx context.Context, mtr repository.Store, serverURL string, client *http.Client, key string) {
	getContext, getCancel := context.WithTimeout(ctx, storeTimeout)
	defer getCancel()

	metricsMap, err := mtr.GetMetrics(getContext)
	if err != nil {
		log.Error().Err(err).Msg("Some error occurred during metrics get")
	}

	serverURL = strings.TrimSuffix(serverURL, "/")
	updateURL := fmt.Sprintf("%s/", serverURL)
	for _, v := range metricsMap {
		err := sendMetricJSON(ctx, updateURL, client, v, key)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to send metric %s", v.ID)
		}
	}
}

func SendBatchJSON(ctx context.Context, mtr repository.Store, serverURL string, client *http.Client) {
	getContext, getCancel := context.WithTimeout(ctx, storeTimeout)
	defer getCancel()

	metricsMap, err := mtr.GetMetrics(getContext)
	if err != nil {
		log.Error().Err(err).Msg("Some error occurred during metrics get")
	}

	metricsSlice := make([]*metrics.Metric, 0)
	for _, v := range metricsMap {
		metricsSlice = append(metricsSlice, v)
	}

	serverURL = strings.TrimSuffix(serverURL, "/")
	updateURL := fmt.Sprintf("%s/updates/", serverURL)

	if err := sendBatchJSON(ctx, updateURL, client, metricsSlice); err != nil {
		log.Error().Err(err).Msg("Filed to send metrics")
	}
}

func sendMetric(ctx context.Context, metricUpdateURL string, client *http.Client) error {
	log.Info().Msgf("Update metric: %s", metricUpdateURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, metricUpdateURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server response: %s", resp.Status)
	}
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
}

func sendMetricJSON(ctx context.Context, serverURL string,
	client *http.Client, metric *metrics.Metric, key string) error {
	log.Info().Msgf("Update metric: %s", metric.ID)

	metric.SetHash(key)
	body, err := metric.EncodeMetric()
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server response: %s", resp.Status)
	}
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
}

func sendBatchJSON(ctx context.Context, metricsUpdateURL string, client *http.Client, metrics []*metrics.Metric) error {
	var buf bytes.Buffer
	jsonEncoder := json.NewEncoder(&buf)

	if err := jsonEncoder.Encode(metrics); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, metricsUpdateURL, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server response: %s", resp.Status)
	}
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
}

func resetCounters(ctx context.Context, mtr repository.Store) {
	resetContext, resetCancel := context.WithTimeout(ctx, storeTimeout)
	defer resetCancel()

	if err := mtr.ResetCounterMetric(resetContext, "PollCount"); err != nil {
		log.Error().Err(err).Msg("couldn't reset counter")
	}
}
