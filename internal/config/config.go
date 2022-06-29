package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	"github.com/itd27m01/go-metrics-service/internal/agent"
	"github.com/itd27m01/go-metrics-service/internal/server"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

const (
	defaultConfigPath = "config.yaml"
)

// Config collects configuration for project
type Config struct {
	Path         string
	ServerConfig server.ServerConfig `yaml:"server"`
	AgentConfig  agent.AgentConfig   `yaml:"agent"`
}

// ParseConfig parses config from file
func (c *Config) ParseConfig(path string) error {
	if path == "" {
		path = defaultConfigPath
	}

	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("can't open config: %w", err)
	}

	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return fmt.Errorf("can't decode yaml config file: %w", err)
	}

	return nil
}

// MergeConfig merges values from flags, file and env
func (c *Config) MergeConfig() {
	pflag.Parse()

	err := c.ParseConfig(c.Path)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config file")
	}

	if err := env.Parse(c); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse environment variables")
	}
}
