package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"

	pb "github.com/itd27m01/go-metrics-service/internal/proto" // import protobufs
	"github.com/itd27m01/go-metrics-service/internal/repository"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

// Config is a config for grpc server
type Config struct {
	Address string `yaml:"address" env:"GRPC_ADDRESS"`
}

// Server implements GRPC server for metrics
type Server struct {
	Cfg          *Config
	SignKey      string
	metricsStore repository.Store
	pb.UnimplementedMetricsServer
}

// Start starts grpc server for metrics
func (s *Server) Start(ctx context.Context, storage repository.Store) error {
	s.metricsStore = storage

	log.Info().Msgf("Start grpcListener on %s", s.Cfg.Address)
	listen, err := net.Listen("tcp", s.Cfg.Address)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start GRPC server")
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMetricsServer(grpcServer, s)
	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	return grpcServer.Serve(listen)
}
