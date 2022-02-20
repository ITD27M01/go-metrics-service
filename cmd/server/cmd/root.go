package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultServerAddress = "127.0.0.1:8080"
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
	defaultStoreInterval = 300
)

var (
	rootCmd = &cobra.Command{
		Use:   "server",
		Short: "Simple metrics server for learning purposes",
		Long:  `Start the server and enjoy a lot of metrics!`,
	}
	ServerAddress string
	Restore       bool
	StoreInterval time.Duration
	StoreFile     string
)

func init() {
	rootCmd.Flags().StringVarP(&ServerAddress, "address", "a", defaultServerAddress,
		"Pair of ip:port to listen on")

	rootCmd.Flags().StringVarP(&StoreFile, "file", "f", defaultStoreFile,
		"Number of seconds to periodically save metrics")

	rootCmd.Flags().BoolVarP(&Restore, "restore", "r", true,
		"Flag to load initial metrics from storage backend")

	rootCmd.Flags().DurationVarP(&StoreInterval, "interval", "i", defaultStoreInterval*time.Second,
		"Number of seconds to periodically save metrics")
}

func Execute() error {
	return rootCmd.Execute()
}
