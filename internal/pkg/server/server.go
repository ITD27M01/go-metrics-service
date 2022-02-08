package server

import (
	"context"
	"fmt"
	"net/http"
)

type Config struct {
	ServerAddress string
	ServerPort    string
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
	registerHandlers(mux)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.Cfg.ServerAddress, s.Cfg.ServerPort),
		Handler: mux,
	}

	s.listener = httpServer

	if err := s.listener.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Printf("HTTP server ListenAndServe: %v", err)
	}
	fmt.Println("HTTP server ListenAndServe exit")
}

func (s *MetricsServer) StopListener() {
	err := s.listener.Shutdown(s.context)
	if err != nil {
		fmt.Printf("HTTP server ListenAndServe shutdown error: %v", err)
	}
}
