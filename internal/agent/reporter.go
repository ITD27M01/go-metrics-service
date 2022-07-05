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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/itd27m01/go-metrics-service/internal/models/metrics"
	pb "github.com/itd27m01/go-metrics-service/internal/proto" // import protobufs
	"github.com/itd27m01/go-metrics-service/internal/repository"
	"github.com/itd27m01/go-metrics-service/pkg/encryption"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
	"github.com/itd27m01/go-metrics-service/pkg/security"
)

// ReporterConfig is a config for reporter worker
type ReporterConfig struct {
	ServerScheme      string        `yaml:"server_scheme" env:"SERVER_SCHEME" envDefault:"http"`
	ServerAddress     string        `yaml:"server_address" env:"ADDRESS"`
	GRPCServerAddress string        `yaml:"grpc_server_address" env:"GRPC_ADDRESS"`
	ServerPath        string        `yaml:"server_path" env:"SERVER_PATH" envDefault:"/update/"`
	ReportInterval    time.Duration `yaml:"report_interval" env:"REPORT_INTERVAL"`
	ServerTimeout     time.Duration `yaml:"server_timeout" env:"SERVER_TIMEOUT"`
	CryptoKey         string        `yaml:"crypto_key" env:"CRYPTO_KEY"`
	SignKey           string        `yaml:"sign_key" env:"KEY"`
}

// ReportWorker defines reporter worker object
type ReportWorker struct {
	Cfg *ReporterConfig
}

// Run runs reporter worker
func (rw *ReportWorker) Run(ctx context.Context, mtr repository.Store) {
	reportTicker := time.NewTicker(rw.Cfg.ReportInterval)
	defer reportTicker.Stop()

	httpClient := rw.getHTTPClient()
	grpcClient, grpcConnection := rw.getGRPCClient()

	serverHTTPURL := rw.Cfg.ServerScheme + "://" + rw.Cfg.ServerAddress
	sendHTTPURL := serverHTTPURL + rw.Cfg.ServerPath

	for {
		select {
		case <-ctx.Done():
			if err := grpcConnection.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close grpc connection")
			}
			return
		case <-reportTicker.C:
			SendHTTPReport(ctx, mtr, sendHTTPURL, httpClient)
			SendHTTPReportJSON(ctx, mtr, sendHTTPURL, httpClient, rw.Cfg.SignKey)
			SendHTTPBatchJSON(ctx, mtr, serverHTTPURL, httpClient)
			SendGRPCReport(ctx, mtr, grpcClient)
			resetCounters(ctx, mtr)
		}
	}
}

// getHTTPClient returns http client
func (rw *ReportWorker) getHTTPClient() *http.Client {
	publicKey, err := encryption.ReadPublicKey(rw.Cfg.CryptoKey)
	if err != nil {
		log.Fatal().Err(err).Msgf("Couldn't read public key from %s", rw.Cfg.CryptoKey)
	}

	transport := http.DefaultTransport
	transport = encryption.NewEncryptRoundTripper(transport, publicKey)
	transport = security.NewRealIPRoundTripper(transport)
	return &http.Client{
		Timeout:   rw.Cfg.ServerTimeout,
		Transport: transport,
	}
}

// getGRPCClient returns grpc client
func (rw *ReportWorker) getGRPCClient() (pb.MetricsClient, *grpc.ClientConn) {
	conn, err := grpc.Dial(rw.Cfg.GRPCServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Msgf("Couldn't create grpc connection %s", rw.Cfg.GRPCServerAddress)
	}

	return pb.NewMetricsClient(conn), conn
}

// SendGRPCReport makes work for sending each metric in url params
func SendGRPCReport(ctx context.Context, mtr repository.Store, client pb.MetricsClient) {
	getContext, getCancel := context.WithTimeout(ctx, pollTimeout)
	defer getCancel()

	metricsMap, err := mtr.GetMetrics(getContext)
	if err != nil {
		log.Error().Err(err).Msgf("Some error occurred during metrics get")
	}

	if len(metricsMap) == 0 {
		return
	}
	stream, err := client.UpdateMetrics(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("Couldn't open grpc stream")
		return
	}
	defer func() {
		if resp, err := stream.CloseAndRecv(); err != nil {
			log.Error().Err(err).Msg("Failed to close stream")
		} else {
			log.Info().Msgf("Close stream: %s", resp.Error)
		}
	}()

	for _, v := range metricsMap {
		request := pb.UpdateMetricRequest{
			Metric: &pb.Metric{
				ID:   v.ID,
				Type: v.MType,
				Hash: v.Hash,
			},
		}
		if v.MType == metrics.MetricTypeGauge {
			request.Metric.Value = float32(*v.Value)
		} else {
			request.Metric.Delta = int64(*v.Delta)
		}
		if err := stream.Send(&request); err != nil {
			log.Error().Err(err).Msgf("Failed to send metric %s", v.ID)
		}
	}
}

// SendHTTPReport makes work for sending each metric in url params
func SendHTTPReport(ctx context.Context, mtr repository.Store, serverURL string, client *http.Client) {
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
		err := sendHTTPMetric(ctx, metricUpdateURL, client)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to send metric %s", v.ID)
		}
	}
}

// SendHTTPReportJSON gets metrics from underlying storage and sends each as a json object
func SendHTTPReportJSON(ctx context.Context, mtr repository.Store, serverURL string, client *http.Client, key string) {
	getContext, getCancel := context.WithTimeout(ctx, pollTimeout)
	defer getCancel()

	metricsMap, err := mtr.GetMetrics(getContext)
	if err != nil {
		log.Error().Err(err).Msg("Some error occurred during metrics get")
	}

	serverURL = strings.TrimSuffix(serverURL, "/")
	updateURL := fmt.Sprintf("%s/", serverURL)
	for _, v := range metricsMap {
		err := sendHTTPMetricJSON(ctx, updateURL, client, v, key)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to send metric %s", v.ID)
		}
	}
}

// SendHTTPBatchJSON gets metrics from underlying storage and sends them as a json list
func SendHTTPBatchJSON(ctx context.Context, mtr repository.Store, serverURL string, client *http.Client) {
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

	if err := sendHTTPBatchJSON(ctx, updateURL, client, metricsSlice); err != nil {
		log.Error().Err(err).Msg("Filed to send metrics")
	}
}

// sendHTTPMetric sends metric in url params
func sendHTTPMetric(ctx context.Context, metricUpdateURL string, client *http.Client) error {
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

// sendHTTPMetricJSON reports to the server a metric in json
func sendHTTPMetricJSON(ctx context.Context, serverURL string,
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

// sendHTTPBatchJSON reports to the server a batch of metrics
func sendHTTPBatchJSON(ctx context.Context, metricsUpdateURL string, client *http.Client, metrics []*metrics.Metric) error {

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
