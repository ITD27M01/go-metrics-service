package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/pkg/repository"
)

type Config struct {
	ServerAddress string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`

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

	initStore(s.Cfg)
	preserverContext, preserverCancel := context.WithCancel(ctx)

	go runPreserver(preserverContext, s.Cfg.MetricsStore, s.Cfg.Restore)

	go s.startListener()
	log.Printf("Start listener on %s", s.Cfg.ServerAddress)

	signalChannel := getSignalChannel()
	signalName := <-signalChannel
	log.Printf("%s signal received, graceful shutdown the server", signalName)
	s.stopListener()

	preserverCancel()

	if err := s.Cfg.MetricsStore.Close(); err != nil {
		log.Printf("Could not close filestore file: %q", err)
	}

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
