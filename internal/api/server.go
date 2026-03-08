package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/nagicantsleep/k-map/internal/config"
)

// NewServer builds the base HTTP server configuration.
func NewServer(cfg config.HTTPConfig, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}

// Run starts the HTTP server and shuts it down when the context is canceled.
func Run(ctx context.Context, server *http.Server, shutdownTimeout time.Duration) error {
	serverErr := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			serverErr <- nil

			return
		}

		serverErr <- err
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return <-serverErr
	}
}
