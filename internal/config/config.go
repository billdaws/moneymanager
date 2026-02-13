package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server    ServerConfig
	Kreuzberg KreuzbergConfig
	Database  DatabaseConfig
	Upload    UploadConfig
	Logging   LoggingConfig
	GnuCash   GnuCashConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// KreuzbergConfig holds Kreuzberg service configuration
type KreuzbergConfig struct {
	URL     string
	Timeout time.Duration
}

// DatabaseConfig holds database paths
type DatabaseConfig struct {
	GnuCashPath  string
	MetadataPath string
}

// UploadConfig holds file upload configuration
type UploadConfig struct {
	MaxSizeMB     int
	AllowedTypes  []string
	TempDir       string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// GnuCashConfig holds GNU Cash specific configuration
type GnuCashConfig struct {
	DefaultCurrency    string
	AutoCreateAccounts bool
}

// Load reads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 3000),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 60*time.Second),
		},
		Kreuzberg: KreuzbergConfig{
			URL:     getEnv("KREUZBERG_URL", "http://localhost:8080"),
			Timeout: getEnvDuration("KREUZBERG_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			GnuCashPath:  getEnv("GNUCASH_DB_PATH", "./data/finance.gnucash"),
			MetadataPath: getEnv("METADATA_DB_PATH", "./data/metadata.db"),
		},
		Upload: UploadConfig{
			MaxSizeMB:    getEnvInt("UPLOAD_MAX_SIZE_MB", 50),
			AllowedTypes: []string{"application/pdf", "text/csv", "application/vnd.ms-excel"},
			TempDir:      getEnv("UPLOAD_TEMP_DIR", "./uploads"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		GnuCash: GnuCashConfig{
			DefaultCurrency:    getEnv("GNUCASH_DEFAULT_CURRENCY", "USD"),
			AutoCreateAccounts: getEnvBool("GNUCASH_AUTO_CREATE_ACCOUNTS", true),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Upload.MaxSizeMB < 1 {
		return fmt.Errorf("invalid upload max size: %d", c.Upload.MaxSizeMB)
	}

	if c.Kreuzberg.URL == "" {
		return fmt.Errorf("kreuzberg URL is required")
	}

	return nil
}

// Helper functions for environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
