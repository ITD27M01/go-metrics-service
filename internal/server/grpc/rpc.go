package grpc

import (
	"errors"
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

		protoMetric := message.Metric
		if protoMetric.Type == metrics.MetricTypeGauge {
			gaugeValue := metrics.Gauge(protoMetric.Value)
			metric = metrics.Metric{
				ID:    protoMetric.ID,
				MType: protoMetric.Type,
				Value: &gaugeValue,
				Hash:  protoMetric.Hash,
			}
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
