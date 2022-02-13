package workers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

type ReporterConfig struct {
	ServerURL      string
	ServerTimeout  time.Duration
	ReportInterval time.Duration
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

	for {
		select {
		case <-reporterContext.Done():
			return
		case <-reportTicker.C:
			SendReport(reporterContext, mtr, rw.Cfg.ServerURL, &client)
		}
	}
}

func SendReport(ctx context.Context, mtr metrics.Store, serverURL string, client *http.Client) {
	for k, v := range mtr.GetGaugeMetrics() {
		metricUpdateURL := fmt.Sprintf("%s/%s/%s/%f", serverURL, metrics.GaugeMetricTypeName, k, v)
		err := updateMetric(ctx, metricUpdateURL, client)
		if err != nil {
			log.Println(err)
		}
	}

	for k, v := range mtr.GetCounterMetrics() {
		metricUpdateURL := fmt.Sprintf("%s/%s/%s/%d", serverURL, metrics.CounterMetricTypeName, k, v)
		err := updateMetric(ctx, metricUpdateURL, client)
		if err != nil {
			log.Println(err)
		}
	}
}

func updateMetric(ctx context.Context, metricUpdateURL string, client *http.Client) error {
	log.Printf("Update metric: %s", metricUpdateURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, metricUpdateURL, nil)
	if err != nil {
		log.Println(err)

		return err
	}
	req.Header.Set("content-type", "text/plain")

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
