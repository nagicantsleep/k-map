package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nagicantsleep/k-map/internal/config"
)

func TestReadinessCheckerSucceedsWhenAllDependenciesReachable(t *testing.T) {
	t.Parallel()

	postgresAddress, closePostgres := newTCPListener(t)
	defer closePostgres()

	redisAddress, closeRedis := newTCPListener(t)
	defer closeRedis()

	nominatim := newNominatimStatusServer(t, http.StatusOK, `{"status":0,"message":"OK"}`)
	defer nominatim.Close()

	checker, err := NewReadinessChecker(config.Config{
		Postgres: config.PostgresConfig{
			Address:     postgresAddress,
			DialTimeout: time.Second,
		},
		Redis: config.RedisConfig{
			Address:     redisAddress,
			DialTimeout: time.Second,
		},
		Nominatim: config.NominatimConfig{
			BaseURL:     nominatim.URL,
			DialTimeout: time.Second,
		},
	})
	if err != nil {
		t.Fatalf("NewReadinessChecker() error = %v", err)
	}

	if err := checker.Check(context.Background()); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestReadinessCheckerFailsWhenNominatimStatusEndpointReportsImporting(t *testing.T) {
	t.Parallel()

	postgresAddress, closePostgres := newTCPListener(t)
	defer closePostgres()

	redisAddress, closeRedis := newTCPListener(t)
	defer closeRedis()

	nominatim := newNominatimStatusServer(t, http.StatusOK, `{"status":700,"message":"importing"}`)
	defer nominatim.Close()

	checker, err := NewReadinessChecker(config.Config{
		Postgres: config.PostgresConfig{
			Address:     postgresAddress,
			DialTimeout: time.Second,
		},
		Redis: config.RedisConfig{
			Address:     redisAddress,
			DialTimeout: time.Second,
		},
		Nominatim: config.NominatimConfig{
			BaseURL:     nominatim.URL,
			DialTimeout: time.Second,
		},
	})
	if err != nil {
		t.Fatalf("NewReadinessChecker() error = %v", err)
	}

	if err := checker.Check(context.Background()); err == nil {
		t.Fatal("Check() error = nil, want error")
	}
}

func TestReadinessCheckerUsesBasePathForNominatimStatusCheck(t *testing.T) {
	t.Parallel()

	postgresAddress, closePostgres := newTCPListener(t)
	defer closePostgres()

	redisAddress, closeRedis := newTCPListener(t)
	defer closeRedis()

	requestedPath := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestedPath <- request.URL.RequestURI()
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"status":0,"message":"OK"}`))
	}))
	defer server.Close()

	checker, err := NewReadinessChecker(config.Config{
		Postgres: config.PostgresConfig{
			Address:     postgresAddress,
			DialTimeout: time.Second,
		},
		Redis: config.RedisConfig{
			Address:     redisAddress,
			DialTimeout: time.Second,
		},
		Nominatim: config.NominatimConfig{
			BaseURL:     server.URL + "/nominatim",
			DialTimeout: time.Second,
		},
	})
	if err != nil {
		t.Fatalf("NewReadinessChecker() error = %v", err)
	}

	if err := checker.Check(context.Background()); err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	gotPath := <-requestedPath
	if gotPath != "/nominatim/status?format=json" {
		t.Fatalf("requested status path = %q, want %q", gotPath, "/nominatim/status?format=json")
	}
}

func TestReadinessCheckerFailsWhenTCPDependencyUnavailable(t *testing.T) {
	t.Parallel()

	postgresAddress, closePostgres := newTCPListener(t)
	defer closePostgres()

	nominatim := newNominatimStatusServer(t, http.StatusOK, `{"status":0,"message":"OK"}`)
	defer nominatim.Close()

	checker, err := NewReadinessChecker(config.Config{
		Postgres: config.PostgresConfig{
			Address:     postgresAddress,
			DialTimeout: time.Second,
		},
		Redis: config.RedisConfig{
			Address:     "127.0.0.1:1",
			DialTimeout: 100 * time.Millisecond,
		},
		Nominatim: config.NominatimConfig{
			BaseURL:     nominatim.URL,
			DialTimeout: time.Second,
		},
	})
	if err != nil {
		t.Fatalf("NewReadinessChecker() error = %v", err)
	}

	if err := checker.Check(context.Background()); err == nil {
		t.Fatal("Check() error = nil, want error")
	}
}

func newTCPListener(t *testing.T) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			connection, acceptErr := listener.Accept()
			if acceptErr != nil {
				return
			}

			_ = connection.Close()
		}
	}()

	return listener.Addr().String(), func() {
		_ = listener.Close()
		<-done
	}
}

func newNominatimStatusServer(t *testing.T, statusCode int, body string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/status" {
			http.NotFound(writer, request)

			return
		}

		if got := request.URL.Query().Get("format"); got != "json" {
			http.Error(writer, fmt.Sprintf("unexpected format query: %q", got), http.StatusBadRequest)

			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(statusCode)
		_, _ = writer.Write([]byte(body))
	}))
}
