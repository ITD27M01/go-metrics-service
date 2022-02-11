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

const metricsHTML = `<!DOCTYPE html>
<html lang="en">
<body>
<table>
    <tr>
        <th>Type</th>
        <th>Name</th>
        <th>Value</th>
    </tr>
    <tr>
        <td style='text-align:center; vertical-align:middle'>Gauge</td>
        <td style='text-align:center; vertical-align:middle'>test</td>
        <td style='text-align:center; vertical-align:middle'>100</td>
    </tr>
    <tr>
        <td style='text-align:center; vertical-align:middle'>Gauge</td>
        <td style='text-align:center; vertical-align:middle'>testSetGet134</td>
        <td style='text-align:center; vertical-align:middle'>96969.519</td>
    </tr>
    <tr>
        <td style='text-align:center; vertical-align:middle'>Gauge</td>
        <td style='text-align:center; vertical-align:middle'>testSetGet135</td>
        <td style='text-align:center; vertical-align:middle'>156519.255</td>
    </tr>
    <tr>
        <td style='text-align:center; vertical-align:middle'>Counter</td>
        <td style='text-align:center; vertical-align:middle'>test</td>
        <td style='text-align:center; vertical-align:middle'>100</td>
    </tr>
    </table>
</body>
</html>`

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
		name:   "Test gauge post 1",
		metric: "/update/gauge/testSetGet134/96969.519",
		method: http.MethodPost,
		want: want{
			code: http.StatusOK,
		},
	},
	{
		name:   "Test gauge post 2",
		metric: "/update/gauge/testSetGet135/156519.255",
		method: http.MethodPost,
		want: want{
			code: http.StatusOK,
		},
	},
	{
		name:   "Test gauge get 1",
		metric: "/value/gauge/testSetGet134",
		method: http.MethodGet,
		want: want{
			code: http.StatusOK,
			data: "96969.519",
		},
	},
	{
		name:   "Test gauge get 2",
		metric: "/value/gauge/testSetGet135",
		method: http.MethodGet,
		want: want{
			code: http.StatusOK,
			data: "156519.255",
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
			data: metricsHTML,
		},
	},
	{
		name:   "Get gauge metric",
		metric: "/value/gauge/test",
		method: http.MethodGet,
		want: want{
			code: http.StatusOK,
			data: "100",
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

	assert.Equal(t, testData.want.code, resp.StatusCode)
	require.NoError(t, err)
	defer resp.Body.Close()

	if testData.method == http.MethodGet {
		respBody, err := ioutil.ReadAll(resp.Body)
		assert.Equal(t, testData.want.data, string(respBody))
		require.NoError(t, err)
	}
}
