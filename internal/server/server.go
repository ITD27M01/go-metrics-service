package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/pkg/logging/log"
	"github.com/itd27m01/go-metrics-service/internal/repository"
)

type Config struct {
	ServerAddress string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFilePath string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	SignKey       string        `env:"KEY"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
	LogLevel      string        `env:"LogLevel"`

	MetricsStore repository.Store
}

type MetricsServer struct {
	Cfg      *Config
	context  context.Context
	listener *http.Server
}

func (s *MetricsServer) Start(ctx context.Context) {
	serverContext, serverCancel := context.WithCancel(ctx)
	s.context = serverContext

	storeContext, storeCancel := context.WithCancel(ctx)

	closeStore := runStore(storeContext, s.Cfg)

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

func getSignalChannel() chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	return signalChannel
}
