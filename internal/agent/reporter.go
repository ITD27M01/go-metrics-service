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

	"github.com/itd27m01/go-metrics-service/internal/models/metrics"
	"github.com/itd27m01/go-metrics-service/internal/repository"
	"github.com/itd27m01/go-metrics-service/pkg/encryption"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
	"github.com/itd27m01/go-metrics-service/pkg/security"
)

// ReporterConfig is a config for reporter worker
type ReporterConfig struct {
	ServerScheme   string        `yaml:"server_scheme" env:"SERVER_SCHEME" envDefault:"http"`
	ServerAddress  string        `yaml:"server_address" env:"ADDRESS"`
	ServerPath     string        `yaml:"server_path" env:"SERVER_PATH" envDefault:"/update/"`
	ReportInterval time.Duration `yaml:"report_interval" env:"REPORT_INTERVAL"`
	ServerTimeout  time.Duration `yaml:"server_timeout" env:"SERVER_TIMEOUT"`
	CryptoKey      string        `yaml:"crypto_key" env:"CRYPTO_KEY"`
	SignKey        string        `yaml:"sign_key" env:"KEY"`
}

// ReportWorker defines reporter worker object
type ReportWorker struct {
	Cfg *ReporterConfig
}

// Run runs reporter worker
func (rw *ReportWorker) Run(ctx context.Context, mtr repository.Store) {
	reportTicker := time.NewTicker(rw.Cfg.ReportInterval)
	defer reportTicker.Stop()

	publicKey, err := encryption.ReadPublicKey(rw.Cfg.CryptoKey)
	if err != nil {
		log.Fatal().Err(err).Msgf("Couldn't read public key from %s", rw.Cfg.CryptoKey)
	}

	transport := http.DefaultTransport
	transport = encryption.NewEncryptRoundTripper(transport, publicKey)
	transport = security.NewRealIPRoundTripper(transport)
	client := http.Client{
		Timeout:   rw.Cfg.ServerTimeout,
		Transport: transport,
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

// SendReport makes work for sending each metric in url params
func SendReport(ctx context.Context, mtr repository.Store, serverURL string, client *http.Client) {
	getContext, getCancel := context.WithTimeout(ctx, pollTimeout)
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

// SendReportJSON gets metrics from underlying storage and sends each as a json object
func SendReportJSON(ctx context.Context, mtr repository.Store, serverURL string, client *http.Client, key string) {
	getContext, getCancel := context.WithTimeout(ctx, pollTimeout)
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

// SendBatchJSON gets metrics from underlying storage and sends them as a json list
func SendBatchJSON(ctx context.Context, mtr repository.Store, serverURL string, client *http.Client) {
	getContext, getCancel := context.WithTimeout(ctx, pollTimeout)
	defer getCancel()

	metricsMap, err := mtr.GetMetrics(getContext)
	if err != nil {
		log.Error().Err(err).Msg("Some error occurred during metrics get")
	}

	metricsSlice := make([]*metrics.Metric, 0, len(metricsMap))
	for _, v := range metricsMap {
		metricsSlice = append(metricsSlice, v)
	}

	serverURL = strings.TrimSuffix(serverURL, "/")
	updateURL := fmt.Sprintf("%s/updates/", serverURL)

	if err := sendBatchJSON(ctx, updateURL, client, metricsSlice); err != nil {
		log.Error().Err(err).Msg("Filed to send metrics")
	}
}

// sendMetric sends metric in url params
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

// sendMetricJSON reports to the server a metric in json
func sendMetricJSON(ctx context.Context, serverURL string,
	client *http.Client, metric *metrics.Metric, key string) error {
	log.Info().Msgf("Update metric: %s", metric.ID)

	metric.SetHash(key)
	body, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL, bytes.NewBuffer(body))
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

// sendBatchJSON reports to the server a batch of metrics
func sendBatchJSON(ctx context.Context, metricsUpdateURL string, client *http.Client, metrics []*metrics.Metric) error {

	encodedMetrics, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, metricsUpdateURL, bytes.NewBuffer(encodedMetrics))
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

// resetCounters helps to reset counter metric
func resetCounters(ctx context.Context, mtr repository.Store) {
	resetContext, resetCancel := context.WithTimeout(ctx, pollTimeout)
	defer resetCancel()

	if err := mtr.ResetCounterMetric(resetContext, "PollCount"); err != nil {
		log.Error().Err(err).Msg("couldn't reset counter")
	}
}
