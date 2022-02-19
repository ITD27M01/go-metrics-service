package workers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

type ReporterConfig struct {
	ServerScheme   string `env:"SERVER_SCHEME" envDefault:"http"`
	ServerAddress  string `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	ServerPath     string `env:"SERVER_PATH" envDefault:"/update/"`
	ServerTimeout  time.Duration
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
}

type ReportWorker struct {
	Cfg ReporterConfig
}

func (rw *ReportWorker) Run(ctx context.Context, mtr metrics.Store) {
	reporterContext, cancel := context.WithCancel(ctx)
	defer cancel()

	reportTicker := time.NewTicker(rw.Cfg.ReportInterval)
	defer reportTicker.Stop()

	client := http.Client{
		Timeout: rw.Cfg.ServerTimeout,
	}

	serverURL := rw.Cfg.ServerScheme + "://" + rw.Cfg.ServerAddress + rw.Cfg.ServerPath

	for {
		select {
		case <-reporterContext.Done():
			return
		case <-reportTicker.C:
			// TODO(igortiunov): Think about storage lock/unlock before/after report
			SendReport(reporterContext, mtr, serverURL, &client)
			SendReportJSON(reporterContext, mtr, serverURL, &client)
			resetCounters(mtr)
		}
	}
}

func SendReport(ctx context.Context, mtr metrics.Store, serverURL string, client *http.Client) {
	serverURL = strings.TrimSuffix(serverURL, "/")
	for k, v := range mtr.GetGaugeMetrics() {
		metricUpdateURL := fmt.Sprintf("%s/%s/%s/%f", serverURL, metrics.GaugeMetricTypeName, k, v)
		err := sendMetric(ctx, metricUpdateURL, client)
		if err != nil {
			log.Println(err)
		}
	}

	for k, v := range mtr.GetCounterMetrics() {
		metricUpdateURL := fmt.Sprintf("%s/%s/%s/%d", serverURL, metrics.CounterMetricTypeName, k, v)
		err := sendMetric(ctx, metricUpdateURL, client)
		if err != nil {
			log.Println(err)
		}
	}
}

func SendReportJSON(ctx context.Context, mtr metrics.Store, serverURL string, client *http.Client) {
	serverURL = strings.TrimSuffix(serverURL, "/")
	updateURL := fmt.Sprintf("%s/", serverURL)
	for k := range mtr.GetGaugeMetrics() {
		v, _ := mtr.GetGaugeMetric(k)
		metric := &metrics.Metric{
			ID:    k,
			MType: metrics.GaugeMetricTypeName,
			Value: &v,
		}
		err := sendMetricJSON(ctx, updateURL, client, metric)
		if err != nil {
			log.Println(err)
		}
	}

	for k := range mtr.GetCounterMetrics() {
		v, _ := mtr.GetCounterMetric(k)
		metric := &metrics.Metric{
			ID:    k,
			MType: metrics.CounterMetricTypeName,
			Delta: &v,
		}
		err := sendMetricJSON(ctx, updateURL, client, metric)
		if err != nil {
			log.Println(err)
		}
	}
}

func sendMetric(ctx context.Context, metricUpdateURL string, client *http.Client) error {
	log.Printf("Update metric: %s", metricUpdateURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, metricUpdateURL, nil)
	if err != nil {
		log.Println(err)

		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)

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

func sendMetricJSON(ctx context.Context, serverURL string, client *http.Client, metric *metrics.Metric) error {
	log.Printf("Update metric: %s", metric.ID)

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

func resetCounters(mtr metrics.Store) {
	mtr.ResetCounterMetric("PollCount")
}
