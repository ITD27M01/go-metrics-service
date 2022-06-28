package server

import (
	"compress/gzip"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/itd27m01/go-metrics-service/pkg/logging"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

// startListener starts http listener for metrics server
func (s *MetricsServer) startListener() {
	mux := chi.NewRouter()

	mux.Use(logging.HTTPRequestLogger(s.Cfg.LogLevel))
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)

	compressor := middleware.NewCompressor(gzip.BestCompression)
	mux.Use(compressor.Handler)

	mux.Mount("/debug", middleware.Profiler())
	RegisterHandlers(mux, s.metricsStore, s.Cfg.SignKey, s.privateKey)

	httpServer := &http.Server{
		Addr:    s.Cfg.ServerAddress,
		Handler: mux,
	}

	s.listener = httpServer

	log.Info().Msgf("%v", s.listener.ListenAndServe())
}

// stopListener stops http listener of metrics server
func (s *MetricsServer) stopListener() {
	err := s.listener.Shutdown(s.context)
	if err != nil {
		log.Info().Msgf("HTTP server ListenAndServe shut down: %v", err)
	}
}
