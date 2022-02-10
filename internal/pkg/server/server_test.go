package server_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	"github.com/itd27m01/go-metrics-service/internal/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const metricsHtml = "<!DOCTYPE html>\n<html lang=\"en\">\n<body>\n<table>\n" +
	"    <tr>\n\t\t<th>Type</th>\n" +
	"        <th>Name</th>\n" +
	"        <th>Value</th>\n" +
	"    </tr>\n    \n" +
	"        <tr>\n\t\t\t<td style='text-align:center; vertical-align:middle'>Gauge</td>\n" +
	"            <td style='text-align:center; vertical-align:middle'>test</td>\n" +
	"            <td style='text-align:center; vertical-align:middle'>100</td>\n" +
	"        </tr>\n    \n    \n" +
	"        <tr>\n\t\t\t<td style='text-align:center; vertical-align:middle'>Counter</td>\n" +
	"            <td style='text-align:center; vertical-align:middle'>test</td>\n" +
	"            <td style='text-align:center; vertical-align:middle'>100</td>\n" +
	"        </tr>\n    \n</table>\n</body>\n</html>\n"

type want struct {
	code int
	data string
}

type test struct {
	name   string
	method string
	metric string
	want   want
}

var tests = []test{
	{
		name:   "OK gauge update",
		metric: "/update/gauge/test/100.000000",
		method: http.MethodPost,
		want: want{
			code: http.StatusOK,
		},
	},
	{
		name:   "OK counter update",
		metric: "/update/counter/test/100",
		method: http.MethodPost,
		want: want{
			code: http.StatusOK,
		},
	},
	{
		name:   "BAD gauge update",
		metric: "/update/gauge/test/none",
		method: http.MethodPost,
		want: want{
			code: http.StatusBadRequest,
		},
	},
	{
		name:   "BAD counter update",
		metric: "/update/counter/test/none",
		method: http.MethodPost,
		want: want{
			code: http.StatusBadRequest,
		},
	},
	{
		name:   "NotFound gauge update",
		metric: "/update/gauge/",
		method: http.MethodPost,
		want: want{
			code: http.StatusNotFound,
		},
	},
	{
		name:   "NotFound counter update",
		metric: "/update/counter/",
		method: http.MethodPost,
		want: want{
			code: http.StatusNotFound,
		},
	},
	{
		name:   "NotImplemented update",
		metric: "/update/unknown/test/1001",
		method: http.MethodPost,
		want: want{
			code: http.StatusNotImplemented,
		},
	},
	{
		name:   "Get all metrics",
		metric: "/",
		method: http.MethodGet,
		want: want{
			code: http.StatusOK,
			data: metricsHtml,
		},
	},
	{
		name:   "Get gauge metric",
		metric: "/value/gauge/test",
		method: http.MethodGet,
		want: want{
			code: http.StatusOK,
			data: "100.000000",
		},
	},
	{
		name:   "Get counter metric",
		metric: "/value/counter/test",
		method: http.MethodGet,
		want: want{
			code: http.StatusOK,
			data: "100",
		},
	},
	{
		name:   "Get unknown metric",
		metric: "/value/counter/unknown",
		method: http.MethodGet,
		want: want{
			code: http.StatusNotFound,
			data: "Metric not found: unknown\n",
		},
	},
}

func TestRouter(t *testing.T) {
	metricsServer := &server.MetricsServer{
		Cfg: server.Config{
			MetricsData: metrics.NewMetrics(),
		}}

	mux := chi.NewRouter()
	server.RegisterHandlers(mux, metricsServer)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRequest(t, ts, tt)
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, testData test) {
	req, err := http.NewRequest(testData.method, ts.URL+testData.metric, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	assert.Equal(t, testData.want.code, resp.StatusCode)
	require.NoError(t, err)

	if testData.method == http.MethodGet {
		respBody, err := ioutil.ReadAll(resp.Body)
		assert.Equal(t, testData.want.data, string(respBody))
		require.NoError(t, err)
	}
}
