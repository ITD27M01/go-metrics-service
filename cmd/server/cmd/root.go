package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultServerAddress = "127.0.0.1:8080"
	defaultStoreFilePath = "/tmp/devops-metrics-db.json"
	defaultStoreInterval = 300 * time.Second
)

var (
	ErrInvalidParam = errors.New("invalid param specified")
	rootCmd         = &cobra.Command{
		Use:   "server",
		Short: "Simple metrics server for learning purposes",
		Long:  `Start the server and enjoy a lot of metrics!`,
		RunE: func(cmd *cobra.Command, args []string) error {
			re := regexp.MustCompile(`(DEBUG|INFO|WARNING|ERROR)`)

			if !re.MatchString(LogLevel) {
				return fmt.Errorf("%w: --log-level", ErrInvalidParam)
			}

			return nil
		},
	}
	ServerAddress string
	Restore       bool
	StoreInterval time.Duration
	StoreFilePath string
	SignKey       string
	DatabaseDSN   string
	LogLevel      string
)

func init() {
	rootCmd.Flags().StringVarP(&ServerAddress, "address", "a", defaultServerAddress,
		"Pair of ip:port to listen on")

	rootCmd.Flags().StringVarP(&StoreFilePath, "file", "f", defaultStoreFilePath,
		"Number of seconds to periodically save metrics")

	rootCmd.Flags().BoolVarP(&Restore, "restore", "r", true,
		"Flag to load initial metrics from storage backend")

	rootCmd.Flags().DurationVarP(&StoreInterval, "interval", "i", defaultStoreInterval,
		"Number of seconds to periodically save metrics")

	rootCmd.Flags().StringVarP(&SignKey, "key", "k", "",
		"Sign key for metrics")

	rootCmd.Flags().StringVarP(&DatabaseDSN, "databaseDSN", "d", "",
		"Database DSN for metrics store")

	rootCmd.Flags().StringVarP(&LogLevel, "log-level", "l", "ERROR",
		"Set log level: DEBUG|INFO|WARNING|ERROR")
}

func Execute() error {
	return rootCmd.Execute()
}
