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
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("KMAP_HTTP_ADDR", ":9090")
	t.Setenv("KMAP_HTTP_READ_HEADER_TIMEOUT", "6s")
	t.Setenv("KMAP_HTTP_READ_TIMEOUT", "11s")
	t.Setenv("KMAP_HTTP_WRITE_TIMEOUT", "16s")
	t.Setenv("KMAP_HTTP_IDLE_TIMEOUT", "61s")
	t.Setenv("KMAP_HTTP_SHUTDOWN_TIMEOUT", "12s")

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
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}
