package grpc

import (
	"errors"
	"fmt"
	"io"

	"github.com/itd27m01/go-metrics-service/internal/models/metrics"
	pb "github.com/itd27m01/go-metrics-service/internal/proto"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

// UpdateMetrics is a stream method updater for metrics
func (s *Server) UpdateMetrics(stream pb.Metrics_UpdateMetricsServer) error {
	log.Info().Msg("GRPC: Start to update metrics")

	var metric metrics.Metric
	metricsSlice := make([]*metrics.Metric, 0)
	for {
		message, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		switch message.Metric.Type {
		case metrics.MetricTypeGauge:
			gaugeValue := metrics.Gauge(message.Metric.Value)
			metric = metrics.Metric{
				ID:    message.Metric.ID,
				MType: message.Metric.Type,
				Value: &gaugeValue,
				Hash:  message.Metric.Hash,
			}
		case metrics.MetricTypeCounter:
			counterValue := metrics.Counter(message.Metric.Delta)
			metric = metrics.Metric{
				ID:    message.Metric.ID,
				MType: message.Metric.Type,
				Delta: &counterValue,
				Hash:  message.Metric.Hash,
			}
		default:
			err := fmt.Errorf("unknown metric type: %s", message.Metric.Type)
			log.Error().Err(err).Msgf("Failed to update metrics")
			return stream.SendAndClose(&pb.UpdateMetricResponse{Error: err.Error()})
		}

		metricsSlice = append(metricsSlice, &metric)
	}

	log.Info().Msgf("GRPC: %d metrics received", len(metricsSlice))

	if err := s.metricsStore.UpdateMetrics(stream.Context(), metricsSlice); err != nil {
		log.Error().Err(err).Msgf("Failed to update metrics")
		return stream.SendAndClose(&pb.UpdateMetricResponse{Error: err.Error()})
	}
	log.Info().Msg("GRPC: Successfully update metrics")

	return stream.SendAndClose(&pb.UpdateMetricResponse{Error: "Metrics are updated"})
}
