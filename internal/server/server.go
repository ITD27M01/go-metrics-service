package server

import (
	"context"
	"crypto/rsa"
	"net/http"
	"os"
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
	context      context.Context
	listener     *http.Server
	metricsStore repository.Store
	privateKey   *rsa.PrivateKey
}

// Start starts a server for metrics collecting
func (s *MetricsServer) Start(ctx context.Context) {
	serverContext, serverCancel := context.WithCancel(ctx)
	s.context = serverContext

	storeContext, storeCancel := context.WithCancel(ctx)

	closeStore := initStore(storeContext, s)

	privateKey, err := encryption.ReadPrivateKey(s.Cfg.CryptoKey)
	if err != nil {
		log.Fatal().Err(err).Msgf("Couldn't read private key from %s", s.Cfg.CryptoKey)
	}
	s.privateKey = privateKey

	go s.startListener()
	log.Info().Msgf("Start listener on %s", s.Cfg.ServerAddress)

	log.Info().Msgf("%s signal received, graceful shutdown the server", <-getSignalChannel())
	s.stopListener()

	if err := closeStore(); err != nil {
		log.Error().Err(err).Msg("Some error occurred while store close")
	}
	storeCancel()

	serverCancel()
}

// getSignalChannel returns a channel from where stop signal will be received
func getSignalChannel() chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	return signalChannel
}
