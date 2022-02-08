package workers

import (
	"context"
	"fmt"
	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	"io"
	"net/http"
	"time"
)

type ReporterConfig struct {
	ServerURL      string
	ServerTimeout  time.Duration
	ReportInterval time.Duration
}

type ReportWorker struct {
	Cfg ReporterConfig
}

func (rw *ReportWorker) Run(ctx context.Context, mtr *metrics.Metrics) {
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
			err := sendReport(reporterContext, mtr, rw.Cfg.ServerURL, client)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func sendReport(ctx context.Context, mtr *metrics.Metrics, serverURL string, client http.Client) (err error) {
	for k, v := range mtr.GaugeMetrics {
		metricUpdateURL := fmt.Sprintf("%s/gauge/%s/%f", serverURL, k, v)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, metricUpdateURL, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
		req.Header.Set("content-type", "text/plain")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		} else {
			_, err = io.Copy(io.Discard, resp.Body)
			if err != nil {
				fmt.Println(err)
			}
			err := resp.Body.Close()
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	for k, v := range mtr.CounterMetrics {
		metricUpdateURL := fmt.Sprintf("%s/counter/%s/%d", serverURL, k, v)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, metricUpdateURL, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
		req.Header.Set("content-type", "text/plain")

		resp, err := client.Do(req)

		if err != nil {
			fmt.Println(err)
		} else {
			_, err = io.Copy(io.Discard, resp.Body)
			if err != nil {
				fmt.Println(err)
			}
			err := resp.Body.Close()
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	return nil
}
