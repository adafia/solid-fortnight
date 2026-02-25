package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration struct.
type Config struct {
	Log      LogConfig                `json:"log" yaml:"log"`
	Storage  StorageConfig            `json:"storage" yaml:"storage"`
	Services map[string]ServiceConfig `json:"services" yaml:"services"`
}

// LogConfig holds the logging configuration.
type LogConfig struct {
	Level  string `json:"level" yaml:"level"`
	Format string `json:"format" yaml:"format"`
}

// StorageConfig holds the configuration for the storage backend.
type StorageConfig struct {
	Type     string         `json:"type" yaml:"type"` // e.g., "postgres", "mysql", "memory"
	Postgres PostgresConfig `json:"postgres,omitempty" yaml:"postgres,omitempty"`
	MySQL    MySQLConfig    `json:"mysql,omitempty" yaml:"mysql,omitempty"`
}

// PostgresConfig holds PostgreSQL connection details.
type PostgresConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	DBName   string `json:"dbname" yaml:"dbname"`
	SSLMode  string `json:"sslmode" yaml:"sslmode"`
}

// MySQLConfig holds MySQL connection details.
type MySQLConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	DBName   string `json:"dbname" yaml:"dbname"`
}

// ServiceConfig holds the configuration for a single service.
type ServiceConfig struct {
	Port int `json:"port" yaml:"port"`
	// Other service-specific settings can be added here.
}

// Load loads the configuration from a YAML file.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &cfg, nil
}
