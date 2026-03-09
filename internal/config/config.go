package config

import (
	"errors"
	"fmt"
	"net/url"
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
	defaultPostgresAddress       = "localhost:5432"
	defaultRedisAddress          = "localhost:6379"
	defaultNominatimBaseURL      = "http://localhost:8081"
	defaultDependencyDialTimeout = 2 * time.Second
)

var errInvalidConfig = errors.New("invalid configuration")

// Config contains process-wide application settings.
type Config struct {
	HTTP      HTTPConfig
	Postgres  PostgresConfig
	Redis     RedisConfig
	Nominatim NominatimConfig
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

// PostgresConfig contains local Postgres dependency settings.
type PostgresConfig struct {
	Address     string
	DialTimeout time.Duration
}

// RedisConfig contains local Redis dependency settings.
type RedisConfig struct {
	Address     string
	DialTimeout time.Duration
}

// NominatimConfig contains local Nominatim dependency settings.
type NominatimConfig struct {
	BaseURL     string
	DialTimeout time.Duration
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
		Postgres: PostgresConfig{
			Address:     getEnv("KMAP_POSTGRES_ADDR", defaultPostgresAddress),
			DialTimeout: getDurationEnv("KMAP_POSTGRES_DIAL_TIMEOUT", defaultDependencyDialTimeout),
		},
		Redis: RedisConfig{
			Address:     getEnv("KMAP_REDIS_ADDR", defaultRedisAddress),
			DialTimeout: getDurationEnv("KMAP_REDIS_DIAL_TIMEOUT", defaultDependencyDialTimeout),
		},
		Nominatim: NominatimConfig{
			BaseURL:     getEnv("KMAP_NOMINATIM_URL", defaultNominatimBaseURL),
			DialTimeout: getDurationEnv("KMAP_NOMINATIM_DIAL_TIMEOUT", defaultDependencyDialTimeout),
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

	if strings.TrimSpace(c.Postgres.Address) == "" {
		return fmt.Errorf("%w: KMAP_POSTGRES_ADDR must not be empty", errInvalidConfig)
	}

	if c.Postgres.DialTimeout <= 0 {
		return fmt.Errorf("%w: KMAP_POSTGRES_DIAL_TIMEOUT must be positive", errInvalidConfig)
	}

	if strings.TrimSpace(c.Redis.Address) == "" {
		return fmt.Errorf("%w: KMAP_REDIS_ADDR must not be empty", errInvalidConfig)
	}

	if c.Redis.DialTimeout <= 0 {
		return fmt.Errorf("%w: KMAP_REDIS_DIAL_TIMEOUT must be positive", errInvalidConfig)
	}

	if err := validateBaseURL(c.Nominatim.BaseURL); err != nil {
		return err
	}

	if c.Nominatim.DialTimeout <= 0 {
		return fmt.Errorf("%w: KMAP_NOMINATIM_DIAL_TIMEOUT must be positive", errInvalidConfig)
	}

	return nil
}

func validateBaseURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("%w: KMAP_NOMINATIM_URL must not be empty", errInvalidConfig)
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: KMAP_NOMINATIM_URL must be a valid URL", errInvalidConfig)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("%w: KMAP_NOMINATIM_URL must use http or https", errInvalidConfig)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("%w: KMAP_NOMINATIM_URL must include a host", errInvalidConfig)
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
