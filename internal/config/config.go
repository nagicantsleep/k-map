package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultHTTPAddress           = ":8080"
	defaultHTTPReadHeaderTimeout = 5 * time.Second
	defaultHTTPReadTimeout       = 10 * time.Second
	defaultHTTPWriteTimeout      = 15 * time.Second
	defaultHTTPIdleTimeout       = 60 * time.Second
	defaultShutdownTimeout       = 10 * time.Second
)

var errInvalidConfig = errors.New("invalid configuration")

// Config contains process-wide application settings.
type Config struct {
	HTTP HTTPConfig
}

// HTTPConfig contains HTTP server settings.
type HTTPConfig struct {
	Address           string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
}

// Load builds Config from environment variables and applies defaults.
func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Address:           getEnv("KMAP_HTTP_ADDR", defaultHTTPAddress),
			ReadHeaderTimeout: getDurationEnv("KMAP_HTTP_READ_HEADER_TIMEOUT", defaultHTTPReadHeaderTimeout),
			ReadTimeout:       getDurationEnv("KMAP_HTTP_READ_TIMEOUT", defaultHTTPReadTimeout),
			WriteTimeout:      getDurationEnv("KMAP_HTTP_WRITE_TIMEOUT", defaultHTTPWriteTimeout),
			IdleTimeout:       getDurationEnv("KMAP_HTTP_IDLE_TIMEOUT", defaultHTTPIdleTimeout),
			ShutdownTimeout:   getDurationEnv("KMAP_HTTP_SHUTDOWN_TIMEOUT", defaultShutdownTimeout),
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate checks for invalid configuration values before startup.
func (c Config) Validate() error {
	if strings.TrimSpace(c.HTTP.Address) == "" {
		return fmt.Errorf("%w: KMAP_HTTP_ADDR must not be empty", errInvalidConfig)
	}

	if c.HTTP.ReadHeaderTimeout <= 0 {
		return fmt.Errorf("%w: KMAP_HTTP_READ_HEADER_TIMEOUT must be positive", errInvalidConfig)
	}

	if c.HTTP.ReadTimeout <= 0 {
		return fmt.Errorf("%w: KMAP_HTTP_READ_TIMEOUT must be positive", errInvalidConfig)
	}

	if c.HTTP.WriteTimeout <= 0 {
		return fmt.Errorf("%w: KMAP_HTTP_WRITE_TIMEOUT must be positive", errInvalidConfig)
	}

	if c.HTTP.IdleTimeout <= 0 {
		return fmt.Errorf("%w: KMAP_HTTP_IDLE_TIMEOUT must be positive", errInvalidConfig)
	}

	if c.HTTP.ShutdownTimeout <= 0 {
		return fmt.Errorf("%w: KMAP_HTTP_SHUTDOWN_TIMEOUT must be positive", errInvalidConfig)
	}

	return nil
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		if strings.TrimSpace(value) == "" {
			return fallback
		}

		return value
	}

	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	if strings.TrimSpace(value) == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}
