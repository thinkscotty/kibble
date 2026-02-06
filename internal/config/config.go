package config

import (
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Logging    LoggingConfig    `yaml:"logging"`
	Similarity SimilarityConfig `yaml:"similarity"`
}

type ServerConfig struct {
	Host                string `yaml:"host"`
	Port                int    `yaml:"port"`
	ReadTimeoutSeconds  int    `yaml:"read_timeout_seconds"`
	WriteTimeoutSeconds int    `yaml:"write_timeout_seconds"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type SimilarityConfig struct {
	Threshold float64 `yaml:"threshold"`
	NGramSize int     `yaml:"ngram_size"`
}

func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Host:                "0.0.0.0",
			Port:                8080,
			ReadTimeoutSeconds:  30,
			WriteTimeoutSeconds: 30,
		},
		Database: DatabaseConfig{
			Path: "./kibble.db",
		},
		Logging: LoggingConfig{
			Level: "info",
		},
		Similarity: SimilarityConfig{
			Threshold: 0.6,
			NGramSize: 3,
		},
	}
}

// Load reads a YAML config file and merges it over defaults.
// If the file does not exist, defaults are returned without error.
func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("No config file found, using defaults", "path", path)
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
