package http

import (
	"compress/gzip"
	"context"
	"crypto/rsa"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/itd27m01/go-metrics-service/internal/repository"
	"github.com/itd27m01/go-metrics-service/pkg/encryption"
	"github.com/itd27m01/go-metrics-service/pkg/logging"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
	"github.com/itd27m01/go-metrics-service/pkg/security"
)

var ErrServerClosed = http.ErrServerClosed

const (
	listenerShutdownTimeout = 30 * time.Second
)

// Config collects configuration for metrics server
type Config struct {
	ServerAddress string `yaml:"address" env:"ADDRESS"`
	CryptoKey     string `yaml:"crypto_key" env:"CRYPTO_KEY"`
	TrustedSubnet string `yaml:"trusted_subnet" env:"TRUSTED_SUBNET"`
}

// Server is a HTTP server for metrics collecting
type Server struct {
	Cfg          *Config
	SignKey      string
	metricsStore repository.Store
	privateKey   *rsa.PrivateKey
}

// Start starts a HTTP server for metrics collecting
func (s *Server) Start(ctx context.Context, storage repository.Store) error {
	s.metricsStore = storage

	privateKey, err := encryption.ReadPrivateKey(s.Cfg.CryptoKey)
	if err != nil {
		log.Fatal().Err(err).Msgf("Couldn't read private key from %s", s.Cfg.CryptoKey)
	}
	s.privateKey = privateKey

	log.Info().Msgf("Start httpListener on %s", s.Cfg.ServerAddress)

	return s.listenAndServe(ctx)
}

// listenAndServe registers handlers and starts listener
func (s *Server) listenAndServe(ctx context.Context) error {
	router := chi.NewRouter()

	router.Use(logging.HTTPRequestLogger())
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(security.CheckRealIP(s.Cfg.TrustedSubnet))
	router.Use(middleware.Recoverer)

	compressor := middleware.NewCompressor(gzip.BestCompression)
	router.Use(compressor.Handler)

	router.Use(encryption.BodyDecrypt(s.privateKey))

	router.Mount("/debug", middleware.Profiler())

	RegisterHandlers(router, s.metricsStore, s.SignKey)
	httpServer := &http.Server{
		Addr:    s.Cfg.ServerAddress,
		Handler: router,
	}
	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), listenerShutdownTimeout)
		defer cancel()
		err := httpServer.Shutdown(ctx)
		if err != nil {
			log.Info().Msgf("HTTP server listenAndServe shut down: %v", err)
		}
	}()

	return httpServer.ListenAndServe()
}
