package server

import (
	"context"
	"crypto/rsa"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/repository"
	"github.com/itd27m01/go-metrics-service/pkg/encryption"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

// ServerConfig collects configuration for metrics server
type ServerConfig struct {
	ServerAddress string        `yaml:"address" env:"ADDRESS"`
	StoreInterval time.Duration `yaml:"store_interval" env:"STORE_INTERVAL"`
	StoreFilePath string        `yaml:"store_file_path" env:"STORE_FILE"`
	Restore       bool          `yaml:"restore" env:"RESTORE"`
	CryptoKey     string        `yaml:"crypto_key" env:"CRYPTO_KEY"`
	SignKey       string        `yaml:"sign_key" env:"KEY"`
	DatabaseDSN   string        `yaml:"database_dsn" env:"DATABASE_DSN"`
	LogLevel      string        `yaml:"log_level" env:"LOG_LEVEL"`
}

// MetricsServer is a server for metrics collecting
type MetricsServer struct {
	Cfg          *ServerConfig
	listener     *http.Server
	metricsStore repository.Store
	privateKey   *rsa.PrivateKey
}

// Start starts a server for metrics collecting
func (s *MetricsServer) Start(parent context.Context) {
	ctx, stop := signal.NotifyContext(parent,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	closeStore := startServerStorage(ctx, s)
	defer func() {
		if err := closeStore(); err != nil {
			log.Error().Err(err).Msg("Some error occurred while store close")
		}
	}()

	privateKey, err := encryption.ReadPrivateKey(s.Cfg.CryptoKey)
	if err != nil {
		log.Fatal().Err(err).Msgf("Couldn't read private key from %s", s.Cfg.CryptoKey)
	}
	s.privateKey = privateKey

	log.Info().Msgf("Start listener on %s", s.Cfg.ServerAddress)
	go s.startListener()
	<-ctx.Done()

	log.Info().Msg("signal received, graceful shutdown the server")
	s.stopListener()
}
