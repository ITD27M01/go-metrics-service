package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

type Config struct {
	ServerAddress string
	ServerPort    string
	MetricsData   *metrics.Metrics
}

type MetricsServer struct {
	Cfg      Config
	context  context.Context
	listener *http.Server
}

func (s *MetricsServer) StartListener(ctx context.Context) {
	serverContext, serverCancel := context.WithCancel(ctx)
	defer serverCancel()

	s.context = serverContext

	mux := http.DefaultServeMux
	registerHandlers(mux, s)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.Cfg.ServerAddress, s.Cfg.ServerPort),
		Handler: mux,
	}

	s.listener = httpServer

	if err := s.listener.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe: %v", err)
	}
	log.Println("HTTP server ListenAndServe exit")
}

func (s *MetricsServer) StopListener() {
	err := s.listener.Shutdown(s.context)
	if err != nil {
		log.Printf("HTTP server ListenAndServe shutdown error: %v", err)
	}
}
