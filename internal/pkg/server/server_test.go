package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	"github.com/itd27m01/go-metrics-service/internal/pkg/server"
)

func TestUserViewHandler(t *testing.T) {
	metricsServer := server.MetricsServer{
		Cfg: server.Config{
			MetricsData: metrics.NewMetrics(),
		}}

	type want struct {
		code int
	}
	tests := []struct {
		name       string
		metricType string
		metric     string
		want       want
	}{
		{
			name:   "OK gauge update",
			metric: "/update/gauge/test/100.0",
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:   "OK counter update",
			metric: "/update/counter/test/100",
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:   "BAD gauge update",
			metric: "/update/gauge/test/none",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:   "BAD counter update",
			metric: "/update/counter/test/none",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:   "NotFound gauge update",
			metric: "/update/gauge/",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "NotFound counter update",
			metric: "/update/counter/",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "NotImplemented update",
			metric: "/update/unknown/test/1001",
			want: want{
				code: http.StatusNotImplemented,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.metric, nil)
			request.Header.Add("Content-Type", "text/plain")

			w := httptest.NewRecorder()
			h := server.UpdateHandler(&metricsServer)
			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			_, err := io.ReadAll(res.Body)
			if err != nil {
				t.Error(err)
			}
			err = res.Body.Close()
			if err != nil {
				return
			}
		})
	}
}
