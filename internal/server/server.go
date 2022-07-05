package server

import (
	"context"
	"errors"
	"os/signal"
	"sync"
	"syscall"

	"github.com/itd27m01/go-metrics-service/internal/config"
	"github.com/itd27m01/go-metrics-service/internal/server/grpc"
	"github.com/itd27m01/go-metrics-service/internal/server/http"
	"github.com/itd27m01/go-metrics-service/internal/server/storage"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

// MetricsServer implements metrics server
type MetricsServer struct {
	Cfg  *config.ServerConfig
	http http.Server
	grpc grpc.Server
}

// Start starts metrics server
func (ms *MetricsServer) Start(parent context.Context) {
	ctx, stop := signal.NotifyContext(parent,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	metricsStorage, closeStore := storage.StartMetricsStorage(ctx, &ms.Cfg.StorageConfig)
	defer func() {
		if err := closeStore(); err != nil {
			log.Error().Err(err).Msg("Some error occurred while store close")
		}
	}()

	wg := sync.WaitGroup{}

	ms.http = http.Server{
		Cfg:     &ms.Cfg.HTTPConfig,
		SignKey: ms.Cfg.SignKey,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := ms.http.Start(ctx, metricsStorage); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msgf("error on listen and serve HTTP server: %s", err)
		}
	}()

	ms.grpc = grpc.Server{
		Cfg:     &ms.Cfg.GRPCConfig,
		SignKey: ms.Cfg.SignKey,
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := ms.grpc.Start(ctx, metricsStorage); err != nil {
			log.Fatal().Err(err).Msgf("error on listen and serve GRPC server: %s", err)
		}
	}()

	wg.Wait()
}
