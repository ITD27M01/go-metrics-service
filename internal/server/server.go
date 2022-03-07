package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/preserver"
	"github.com/itd27m01/go-metrics-service/internal/repository"
)

type Config struct {
	ServerAddress string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFilePath string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	SignKey       string        `env:"KEY"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`

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
	syncChannel := initStore(storeContext, s.Cfg)
	metricsPreserver := preserver.NewPreserver(s.Cfg.MetricsStore, s.Cfg.StoreInterval, syncChannel)
	preserverContext, preserverCancel := context.WithCancel(ctx)

	if s.Cfg.Restore && s.Cfg.MetricsStore.LoadMetrics() != nil {
		log.Println("Filed to load metrics from store")
	}

	go metricsPreserver.RunPreserver(preserverContext)

	go s.startListener()
	log.Printf("Start listener on %s", s.Cfg.ServerAddress)

	log.Printf("%s signal received, graceful shutdown the server", <-getSignalChannel())
	s.stopListener()

	preserverCancel()

	if err := s.Cfg.MetricsStore.Close(); err != nil {
		log.Printf("Could not close filestore: %q", err)
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
