package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/itd27m01/go-metrics-service/internal/models/metrics"

	"github.com/itd27m01/go-metrics-service/internal/repository"
	"github.com/itd27m01/go-metrics-service/internal/server"
)

func ExamplePingHandler() {
	mux := chi.NewRouter()
	server.RegisterHandlers(mux, repository.NewInMemoryStore(), "")
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/ping", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Println(resp.Status)

	// Output: 200 OK
}

func ExampleUpdateHandler() {
	mux := chi.NewRouter()
	server.RegisterHandlers(mux, repository.NewInMemoryStore(), "")
	ts := httptest.NewServer(mux)
	defer ts.Close()

	gaugeValue := metrics.Gauge(96969.519)
	metric := &metrics.Metric{
		ID:    "Alloc",
		MType: metrics.MetricTypeGauge,
		Value: &gaugeValue,
	}

	body, _ := metric.EncodeMetric()
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/update/", body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Printf("By JSON: %s", resp.Status)

	metricURL := ts.URL + "/update/gauge/" + fmt.Sprintf("%s/%s", metric.ID, metric.String())

	req, _ = http.NewRequest(http.MethodPost, metricURL, nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Printf(", By url params: %s", resp.Status)

	// Output: By JSON: 200 OK, By url params: 200 OK
}

func ExampleUpdatesHandler() {
	mux := chi.NewRouter()
	server.RegisterHandlers(mux, repository.NewInMemoryStore(), "")
	ts := httptest.NewServer(mux)
	defer ts.Close()

	gaugeValue := metrics.Gauge(96969.519)
	metricsSlice := []*metrics.Metric{{
		ID:    "Alloc",
		MType: metrics.MetricTypeGauge,
		Value: &gaugeValue,
	}}

	var body bytes.Buffer
	jsonEncoder := json.NewEncoder(&body)

	_ = jsonEncoder.Encode(metricsSlice)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/updates/", &body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Println(resp.Status)

	// Output: 200 OK
}

func ExampleGetMetricHandler() {
	mux := chi.NewRouter()
	server.RegisterHandlers(mux, repository.NewInMemoryStore(), "")
	ts := httptest.NewServer(mux)
	defer ts.Close()

	gaugeValue := metrics.Gauge(96969.519)
	metric := &metrics.Metric{
		ID:    "Alloc",
		MType: metrics.MetricTypeGauge,
		Value: &gaugeValue,
	}

	body, _ := metric.EncodeMetric()
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/update/", body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Println(resp.Status)

	body, _ = metric.EncodeMetric()
	req, _ = http.NewRequest(http.MethodPost, ts.URL+"/value/", body)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Println(resp.Status)

	// Output: 200 OK
	// 200 OK
}
