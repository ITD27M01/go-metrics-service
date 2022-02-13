package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

type Config struct {
	ServerAddress string
	ServerPort    string
	MetricsStore  metrics.Store
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

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)

	RegisterHandlers(mux, s.Cfg.MetricsStore)

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
