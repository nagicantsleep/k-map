package config

import (
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("KMAP_HTTP_ADDR", "")
	t.Setenv("KMAP_HTTP_READ_HEADER_TIMEOUT", "")
	t.Setenv("KMAP_HTTP_READ_TIMEOUT", "")
	t.Setenv("KMAP_HTTP_WRITE_TIMEOUT", "")
	t.Setenv("KMAP_HTTP_IDLE_TIMEOUT", "")
	t.Setenv("KMAP_HTTP_SHUTDOWN_TIMEOUT", "")
	t.Setenv("KMAP_POSTGRES_ADDR", "")
	t.Setenv("KMAP_POSTGRES_DIAL_TIMEOUT", "")
	t.Setenv("KMAP_REDIS_ADDR", "")
	t.Setenv("KMAP_REDIS_DIAL_TIMEOUT", "")
	t.Setenv("KMAP_NOMINATIM_URL", "")
	t.Setenv("KMAP_NOMINATIM_DIAL_TIMEOUT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTP.Address != defaultHTTPAddress {
		t.Fatalf("Address = %q, want %q", cfg.HTTP.Address, defaultHTTPAddress)
	}

	if cfg.HTTP.ReadHeaderTimeout != defaultHTTPReadHeaderTimeout {
		t.Fatalf("ReadHeaderTimeout = %v, want %v", cfg.HTTP.ReadHeaderTimeout, defaultHTTPReadHeaderTimeout)
	}

	if cfg.HTTP.ReadTimeout != defaultHTTPReadTimeout {
		t.Fatalf("ReadTimeout = %v, want %v", cfg.HTTP.ReadTimeout, defaultHTTPReadTimeout)
	}

	if cfg.HTTP.WriteTimeout != defaultHTTPWriteTimeout {
		t.Fatalf("WriteTimeout = %v, want %v", cfg.HTTP.WriteTimeout, defaultHTTPWriteTimeout)
	}

	if cfg.HTTP.IdleTimeout != defaultHTTPIdleTimeout {
		t.Fatalf("IdleTimeout = %v, want %v", cfg.HTTP.IdleTimeout, defaultHTTPIdleTimeout)
	}

	if cfg.HTTP.ShutdownTimeout != defaultShutdownTimeout {
		t.Fatalf("ShutdownTimeout = %v, want %v", cfg.HTTP.ShutdownTimeout, defaultShutdownTimeout)
	}

	if cfg.Postgres.Address != defaultPostgresAddress {
		t.Fatalf("Postgres.Address = %q, want %q", cfg.Postgres.Address, defaultPostgresAddress)
	}

	if cfg.Postgres.DialTimeout != defaultDependencyDialTimeout {
		t.Fatalf("Postgres.DialTimeout = %v, want %v", cfg.Postgres.DialTimeout, defaultDependencyDialTimeout)
	}

	if cfg.Redis.Address != defaultRedisAddress {
		t.Fatalf("Redis.Address = %q, want %q", cfg.Redis.Address, defaultRedisAddress)
	}

	if cfg.Redis.DialTimeout != defaultDependencyDialTimeout {
		t.Fatalf("Redis.DialTimeout = %v, want %v", cfg.Redis.DialTimeout, defaultDependencyDialTimeout)
	}

	if cfg.Nominatim.BaseURL != defaultNominatimBaseURL {
		t.Fatalf("Nominatim.BaseURL = %q, want %q", cfg.Nominatim.BaseURL, defaultNominatimBaseURL)
	}

	if cfg.Nominatim.DialTimeout != defaultDependencyDialTimeout {
		t.Fatalf("Nominatim.DialTimeout = %v, want %v", cfg.Nominatim.DialTimeout, defaultDependencyDialTimeout)
	}
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("KMAP_HTTP_ADDR", ":9090")
	t.Setenv("KMAP_HTTP_READ_HEADER_TIMEOUT", "6s")
	t.Setenv("KMAP_HTTP_READ_TIMEOUT", "11s")
	t.Setenv("KMAP_HTTP_WRITE_TIMEOUT", "16s")
	t.Setenv("KMAP_HTTP_IDLE_TIMEOUT", "61s")
	t.Setenv("KMAP_HTTP_SHUTDOWN_TIMEOUT", "12s")
	t.Setenv("KMAP_POSTGRES_ADDR", "postgres.internal:5432")
	t.Setenv("KMAP_POSTGRES_DIAL_TIMEOUT", "3s")
	t.Setenv("KMAP_REDIS_ADDR", "redis.internal:6379")
	t.Setenv("KMAP_REDIS_DIAL_TIMEOUT", "4s")
	t.Setenv("KMAP_NOMINATIM_URL", "https://nominatim.internal:8080")
	t.Setenv("KMAP_NOMINATIM_DIAL_TIMEOUT", "5s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTP.Address != ":9090" {
		t.Fatalf("Address = %q, want %q", cfg.HTTP.Address, ":9090")
	}

	if cfg.HTTP.ReadHeaderTimeout != 6*time.Second {
		t.Fatalf("ReadHeaderTimeout = %v, want %v", cfg.HTTP.ReadHeaderTimeout, 6*time.Second)
	}

	if cfg.HTTP.ReadTimeout != 11*time.Second {
		t.Fatalf("ReadTimeout = %v, want %v", cfg.HTTP.ReadTimeout, 11*time.Second)
	}

	if cfg.HTTP.WriteTimeout != 16*time.Second {
		t.Fatalf("WriteTimeout = %v, want %v", cfg.HTTP.WriteTimeout, 16*time.Second)
	}

	if cfg.HTTP.IdleTimeout != 61*time.Second {
		t.Fatalf("IdleTimeout = %v, want %v", cfg.HTTP.IdleTimeout, 61*time.Second)
	}

	if cfg.HTTP.ShutdownTimeout != 12*time.Second {
		t.Fatalf("ShutdownTimeout = %v, want %v", cfg.HTTP.ShutdownTimeout, 12*time.Second)
	}

	if cfg.Postgres.Address != "postgres.internal:5432" {
		t.Fatalf("Postgres.Address = %q, want %q", cfg.Postgres.Address, "postgres.internal:5432")
	}

	if cfg.Postgres.DialTimeout != 3*time.Second {
		t.Fatalf("Postgres.DialTimeout = %v, want %v", cfg.Postgres.DialTimeout, 3*time.Second)
	}

	if cfg.Redis.Address != "redis.internal:6379" {
		t.Fatalf("Redis.Address = %q, want %q", cfg.Redis.Address, "redis.internal:6379")
	}

	if cfg.Redis.DialTimeout != 4*time.Second {
		t.Fatalf("Redis.DialTimeout = %v, want %v", cfg.Redis.DialTimeout, 4*time.Second)
	}

	if cfg.Nominatim.BaseURL != "https://nominatim.internal:8080" {
		t.Fatalf("Nominatim.BaseURL = %q, want %q", cfg.Nominatim.BaseURL, "https://nominatim.internal:8080")
	}

	if cfg.Nominatim.DialTimeout != 5*time.Second {
		t.Fatalf("Nominatim.DialTimeout = %v, want %v", cfg.Nominatim.DialTimeout, 5*time.Second)
	}
}

func TestValidateRejectsInvalidConfig(t *testing.T) {
	cfg := Config{
		HTTP: HTTPConfig{
			Address:           " ",
			ReadHeaderTimeout: time.Second,
			ReadTimeout:       time.Second,
			WriteTimeout:      time.Second,
			IdleTimeout:       time.Second,
			ShutdownTimeout:   time.Second,
		},
		Postgres: PostgresConfig{
			Address:     defaultPostgresAddress,
			DialTimeout: time.Second,
		},
		Redis: RedisConfig{
			Address:     defaultRedisAddress,
			DialTimeout: time.Second,
		},
		Nominatim: NominatimConfig{
			BaseURL:     defaultNominatimBaseURL,
			DialTimeout: time.Second,
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestValidateRejectsInvalidNominatimURL(t *testing.T) {
	cfg := Config{
		HTTP: HTTPConfig{
			Address:           defaultHTTPAddress,
			ReadHeaderTimeout: time.Second,
			ReadTimeout:       time.Second,
			WriteTimeout:      time.Second,
			IdleTimeout:       time.Second,
			ShutdownTimeout:   time.Second,
		},
		Postgres: PostgresConfig{
			Address:     defaultPostgresAddress,
			DialTimeout: time.Second,
		},
		Redis: RedisConfig{
			Address:     defaultRedisAddress,
			DialTimeout: time.Second,
		},
		Nominatim: NominatimConfig{
			BaseURL:     "postgres:8080",
			DialTimeout: time.Second,
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}
