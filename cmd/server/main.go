package main

import (
	"context"
	"os"
	"time"

	"github.com/spf13/pflag"

	"github.com/itd27m01/go-metrics-service/internal/config"
	"github.com/itd27m01/go-metrics-service/internal/greetings"
	"github.com/itd27m01/go-metrics-service/internal/server"
	"github.com/itd27m01/go-metrics-service/pkg/logging"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

const (
	defaultServerAddress = "127.0.0.1:8080"
	defaultStoreFilePath = "/tmp/devops-metrics-db.json"
	defaultStoreInterval = 300 * time.Second
)

var (
	Config config.Config
)

func init() {
	pflag.StringVarP(&Config.Path, "config", "c", os.Getenv("CONFIG"),
		"Server config file path")

	pflag.StringVarP(&Config.ServerConfig.ServerAddress, "address", "a", defaultServerAddress,
		"Pair of ip:port to listen on")

	pflag.StringVarP(&Config.ServerConfig.StoreFilePath, "file", "f", defaultStoreFilePath,
		"Number of seconds to periodically save metrics")

	pflag.BoolVarP(&Config.ServerConfig.Restore, "restore", "r", false,
		"Flag to load initial metrics from storage backend")

	pflag.DurationVarP(&Config.ServerConfig.StoreInterval, "interval", "i", defaultStoreInterval,
		"Number of seconds to periodically save metrics")

	pflag.StringVar(&Config.ServerConfig.CryptoKey, "crypto-key", "",
		"A path to the pem file of private RSA key")

	pflag.StringVarP(&Config.ServerConfig.SignKey, "key", "k", "",
		"Sign key for metrics")

	pflag.StringVarP(&Config.ServerConfig.DatabaseDSN, "databaseDSN", "d", "",
		"Database DSN for metrics store")

	pflag.StringVarP(&Config.ServerConfig.LogLevel, "log-level", "l", "ERROR",
		"Set log level: DEBUG|INFO|WARNING|ERROR")
}

func main() {
	Config.MergeConfig()

	logging.LogLevel(Config.ServerConfig.LogLevel)

	metricsServer := server.MetricsServer{Cfg: &Config.ServerConfig}

	if err := greetings.Print(buildVersion, buildDate, buildCommit); err != nil {
		log.Fatal().Err(err).Msg("Failed to start agent, failed to print greetings")
	}

	metricsServer.Start(context.Background())
}
