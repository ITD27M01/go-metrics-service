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
	ServerUrl      string
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
			err := sendReport(reporterContext, mtr, rw.Cfg.ServerUrl, client)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func sendReport(ctx context.Context, mtr *metrics.Metrics, serverUrl string, client http.Client) (err error) {
	for k, v := range mtr.GaugeMetrics {
		metricUpdateUrl := fmt.Sprintf("%s/gauge/%s/%f", serverUrl, k, v)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, metricUpdateUrl, nil)
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
		metricUpdateUrl := fmt.Sprintf("%s/counter/%s/%d", serverUrl, k, v)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, metricUpdateUrl, nil)
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
