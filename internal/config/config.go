package config

import (
	"fmt"
	"os"

	"github.com/itd27m01/go-metrics-service/internal/server/grpc"

	"github.com/itd27m01/go-metrics-service/internal/server/http"
	"github.com/itd27m01/go-metrics-service/internal/server/storage"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"

	"github.com/itd27m01/go-metrics-service/internal/agent"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

// Config collects configuration for project
type Config struct {
	Path         string
	ServerConfig ServerConfig      `yaml:"server"`
	AgentConfig  agent.AgentConfig `yaml:"agent"`
}

type ServerConfig struct {
	HTTPConfig    http.Config    `yaml:"http"`
	GRPCConfig    grpc.Config    `yaml:"grpc"`
	StorageConfig storage.Config `yaml:"storage"`
	SignKey       string         `yaml:"sign_key" env:"KEY"`
	LogLevel      string         `yaml:"log_level" env:"LOG_LEVEL"`
}

// ParseConfig parses config from file
func (c *Config) ParseConfig(path string) error {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("can't open config: %w", err)
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return fmt.Errorf("can't decode yaml config file: %w", err)
	}

	return nil
}

// MergeConfig merges values from flags, file and env
func (c *Config) MergeConfig() {
	err := c.ParseConfig(c.Path)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config file")
	}

	if err := env.Parse(c); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse environment variables")
	}
}
