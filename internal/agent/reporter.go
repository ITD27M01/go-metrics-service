package agent

import (
	"bytes"
	"context"
	"crypto/rsa"
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
)

// ReporterConfig is a config for reporter worker
type ReporterConfig struct {
	ServerScheme   string        `env:"SERVER_SCHEME" envDefault:"http"`
	ServerAddress  string        `env:"ADDRESS"`
	ServerPath     string        `env:"SERVER_PATH" envDefault:"/update/"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	ServerTimeout  time.Duration `env:"SERVER_TIMEOUT"`
	CryptoKey      string        `env:"CRYPTO_KEY"`
	SignKey        string        `env:"KEY"`
}

// ReportWorker defines reporter worker object
type ReportWorker struct {
	Cfg *ReporterConfig
}

// Run runs reporter worker
func (rw *ReportWorker) Run(ctx context.Context, mtr repository.Store) {
	reportTicker := time.NewTicker(rw.Cfg.ReportInterval)
	defer reportTicker.Stop()

	client := http.Client{
		Timeout: rw.Cfg.ServerTimeout,
	}

	serverURL := rw.Cfg.ServerScheme + "://" + rw.Cfg.ServerAddress
	sendURL := serverURL + rw.Cfg.ServerPath

	publicKey, err := encryption.ReadPublicKey(rw.Cfg.CryptoKey)
	if err != nil {
		log.Fatal().Err(err).Msgf("Couldn't read public key from %s", rw.Cfg.CryptoKey)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-reportTicker.C:
			SendReport(ctx, mtr, sendURL, &client)
			SendReportJSON(ctx, mtr, sendURL, &client, rw.Cfg.SignKey, publicKey)
			SendBatchJSON(ctx, mtr, serverURL, &client, publicKey)
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
func SendReportJSON(ctx context.Context, mtr repository.Store, serverURL string, client *http.Client, key string, publicKey *rsa.PublicKey) {
	getContext, getCancel := context.WithTimeout(ctx, pollTimeout)
	defer getCancel()

	metricsMap, err := mtr.GetMetrics(getContext)
	if err != nil {
		log.Error().Err(err).Msg("Some error occurred during metrics get")
	}

	serverURL = strings.TrimSuffix(serverURL, "/")
	updateURL := fmt.Sprintf("%s/", serverURL)
	for _, v := range metricsMap {
		err := sendMetricJSON(ctx, updateURL, client, v, key, publicKey)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to send metric %s", v.ID)
		}
	}
}

// SendBatchJSON gets metrics from underlying storage and sends them as a json list
func SendBatchJSON(ctx context.Context, mtr repository.Store, serverURL string, client *http.Client, publicKey *rsa.PublicKey) {
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

	if err := sendBatchJSON(ctx, updateURL, client, metricsSlice, publicKey); err != nil {
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
	client *http.Client, metric *metrics.Metric, key string, publicKey *rsa.PublicKey) error {
	log.Info().Msgf("Update metric: %s", metric.ID)

	metric.SetHash(key)
	body, err := json.Marshal(metric)
	if err != nil {
		return err
	}
	encryptedBody, err := encryption.RSAEncrypt(body, publicKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL, bytes.NewBuffer(encryptedBody))
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
func sendBatchJSON(ctx context.Context, metricsUpdateURL string, client *http.Client, metrics []*metrics.Metric,
	publicKey *rsa.PublicKey) error {

	encodedMetrics, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	encryptedBody, err := encryption.RSAEncrypt(encodedMetrics, publicKey)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, metricsUpdateURL, bytes.NewBuffer(encryptedBody))
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
