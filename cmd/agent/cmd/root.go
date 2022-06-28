package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultServerAddress  = "127.0.0.1:8080"
	defaultPollInterval   = 2 * time.Second
	defaultReportInterval = 10 * time.Second
	defaultServerTimeout  = 1 * time.Second
)

var (
	ErrInvalidParam = errors.New("invalid param specified")
	rootCmd         = &cobra.Command{
		Use:   "agent",
		Short: "Simple metrics agent for learning purposes",
		Long:  `Start the agent and enjoy a lot of metrics!`,
		RunE: func(cmd *cobra.Command, args []string) error {
			re := regexp.MustCompile(`(DEBUG|INFO|WARNING|ERROR)`)

			if !re.MatchString(LogLevel) {
				return fmt.Errorf("%w: --log-level", ErrInvalidParam)
			}

			return nil
		},
	}
	ServerAddress  string
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerTimeout  time.Duration
	CryptoKey      string
	SignKey        string
	LogLevel       string
)

func init() {
	rootCmd.Flags().StringVarP(&ServerAddress, "address", "a", defaultServerAddress,
		"Pair of ip:port to connect to")

	rootCmd.Flags().DurationVarP(&ServerTimeout, "timeout", "t", defaultServerTimeout,
		"Timeout for server connection")

	rootCmd.Flags().DurationVarP(&ReportInterval, "report", "r", defaultReportInterval,
		"Number of seconds to periodically report metrics")

	rootCmd.Flags().DurationVarP(&PollInterval, "poll", "p", defaultPollInterval,
		"Number of seconds to periodically get metrics")

	rootCmd.Flags().StringVar(&CryptoKey, "crypto-key", "",
		"A path to the pem file of public RSA key")

	rootCmd.Flags().StringVarP(&SignKey, "key", "k", "",
		"Sign key for metrics")

	rootCmd.Flags().StringVarP(&LogLevel, "log-level", "l", "ERROR",
		"Set log level: DEBUG|INFO|WARNING|ERROR")
}

func Execute() error {
	return rootCmd.Execute()
}
