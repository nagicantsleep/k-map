package api

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/nagicantsleep/k-map/internal/config"
)

type dependencyCheck struct {
	address string
	name    string
	timeout time.Duration
}

type networkReadinessChecker struct {
	checks []dependencyCheck
}

// NewReadinessChecker builds a dependency reachability checker from process config.
func NewReadinessChecker(cfg config.Config) (ReadinessChecker, error) {
	nominatimAddress, err := hostPortFromURL(cfg.Nominatim.BaseURL)
	if err != nil {
		return nil, err
	}

	return networkReadinessChecker{
		checks: []dependencyCheck{
			{
				address: cfg.Postgres.Address,
				name:    "postgres",
				timeout: cfg.Postgres.DialTimeout,
			},
			{
				address: cfg.Redis.Address,
				name:    "redis",
				timeout: cfg.Redis.DialTimeout,
			},
			{
				address: nominatimAddress,
				name:    "nominatim",
				timeout: cfg.Nominatim.DialTimeout,
			},
		},
	}, nil
}

func (checker networkReadinessChecker) Check(ctx context.Context) error {
	for _, check := range checker.checks {
		dialer := net.Dialer{Timeout: check.timeout}

		connection, err := dialer.DialContext(ctx, "tcp", check.address)
		if err != nil {
			return fmt.Errorf("dependency %s unreachable at %s: %w", check.name, check.address, err)
		}

		_ = connection.Close()
	}

	return nil
}

func hostPortFromURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	host := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		if parsedURL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	return net.JoinHostPort(host, port), nil
}
