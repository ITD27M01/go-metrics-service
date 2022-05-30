package agent_test

import (
	"context"
	"testing"

	"github.com/itd27m01/go-metrics-service/internal/agent"
	"github.com/itd27m01/go-metrics-service/internal/repository"
)

func TestPoolWorker(t *testing.T) {
	mtr := repository.NewInMemoryStore()
	agent.UpdateMemStatsMetrics(context.Background(), mtr)

	counterMetric, _ := mtr.GetMetric(context.Background(), "PollCount", "")
	if *counterMetric.Delta != 1 {
		t.Errorf("Counter wasn't incremented: %d", *counterMetric.Delta)
	}
}
