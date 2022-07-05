package main

import (
	"context"
	"os"
	"time"

	"github.com/spf13/pflag"

	"github.com/itd27m01/go-metrics-service/internal/agent"
	"github.com/itd27m01/go-metrics-service/internal/config"
	"github.com/itd27m01/go-metrics-service/internal/greetings"
	"github.com/itd27m01/go-metrics-service/pkg/logging"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

const (
	defaultServerScheme      = "http"
	defaultHTTPServerAddress = "127.0.0.1:8080"
	defaultGRPCServerAddress = "127.0.0.1:8081"
	defaultPollInterval      = 2 * time.Second
	defaultReportInterval    = 10 * time.Second
	defaultServerTimeout     = 1 * time.Second
)

var (
	Config config.Config
)

func init() {
	pflag.StringVarP(&Config.Path, "config", "c", os.Getenv("CONFIG"),
		"Agent config file path")

	pflag.StringVar(&Config.AgentConfig.ReporterConfig.ServerAddress, "server-scheme", defaultServerScheme,
		"Server scheme http or https")

	pflag.StringVarP(&Config.AgentConfig.ReporterConfig.ServerAddress, "address", "a", defaultHTTPServerAddress,
		"Pair of ip:port to connect to HTTP server")

	pflag.StringVarP(&Config.AgentConfig.ReporterConfig.GRPCServerAddress, "grpc-address", "g", defaultGRPCServerAddress,
		"Pair of ip:port to connect to GRPC server")

	pflag.DurationVarP(&Config.AgentConfig.ReporterConfig.ServerTimeout, "timeout", "t", defaultServerTimeout,
		"Timeout for server connection")

	pflag.DurationVarP(&Config.AgentConfig.ReporterConfig.ReportInterval, "report", "r", defaultReportInterval,
		"Number of seconds to periodically report metrics")

	pflag.DurationVarP(&Config.AgentConfig.PollerConfig.PollInterval, "poll", "p", defaultPollInterval,
		"Number of seconds to periodically get metrics")

	pflag.StringVar(&Config.AgentConfig.ReporterConfig.CryptoKey, "crypto-key", "",
		"A path to the pem file of public RSA key")

	pflag.StringVarP(&Config.AgentConfig.ReporterConfig.SignKey, "key", "k", "",
		"Sign key for metrics")

	pflag.StringVarP(&Config.AgentConfig.LogLevel, "log-level", "l", "ERROR",
		"Set log level: DEBUG|INFO|WARNING|ERROR")
}

func main() {
	if err := greetings.Print(buildVersion, buildDate, buildCommit); err != nil {
		log.Fatal().Err(err).Msg("Failed to start agent, failed to print greetings")
	}

	pflag.Parse()
	Config.MergeConfig()

	logging.LogLevel(Config.AgentConfig.LogLevel)

	agent.Start(context.Background(), &Config.AgentConfig)
}
