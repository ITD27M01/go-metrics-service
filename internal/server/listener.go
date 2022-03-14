package server

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/rs/zerolog/log"
)

func (s *MetricsServer) startListener() {
	logger := httplog.NewLogger("metrics", httplog.Options{
		JSON:     true,
		LogLevel: strings.ToLower(s.Cfg.LogLevel),
	})

	mux := chi.NewRouter()

	mux.Use(httplog.RequestLogger(logger))
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)

	compressor := middleware.NewCompressor(gzip.BestCompression)
	mux.Use(compressor.Handler)

	RegisterHandlers(mux, s.Cfg.MetricsStore, s.Cfg.SignKey)

	httpServer := &http.Server{
		Addr:    s.Cfg.ServerAddress,
		Handler: mux,
	}

	s.listener = httpServer

	log.Info().Msgf("%v", s.listener.ListenAndServe())
}

func (s *MetricsServer) stopListener() {
	err := s.listener.Shutdown(s.context)
	if err != nil {
		log.Info().Msgf("HTTP server ListenAndServe shut down: %v", err)
	}
}
