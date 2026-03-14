package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/nagicantsleep/k-map/internal/config"
)

// DependencyStatus holds the status of a single dependency.
type DependencyStatus struct {
	Name   string
	Status string
	OK     bool
}

// ReadinessResult holds the overall readiness check result.
type ReadinessResult struct {
	Status       string
	Dependencies map[string]string
}

type dependencyCheck struct {
	address string
	name    string
	timeout time.Duration
}

type networkReadinessChecker struct {
	checks    []dependencyCheck
	nominatim nominatimReadinessCheck
}

type nominatimReadinessCheck struct {
	statusURL string
	timeout   time.Duration
}

type nominatimStatusResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// NewReadinessChecker builds a dependency reachability checker from process config.
func NewReadinessChecker(cfg config.Config) (ReadinessChecker, error) {
	statusURL, err := nominatimStatusURL(cfg.Nominatim.BaseURL)
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
		},
		nominatim: nominatimReadinessCheck{
			statusURL: statusURL,
			timeout:   cfg.Nominatim.DialTimeout,
		},
	}, nil
}

// CheckAll returns the status of all dependencies.
func (checker networkReadinessChecker) CheckAll(ctx context.Context) ReadinessResult {
	result := ReadinessResult{
		Status:       "ok",
		Dependencies: make(map[string]string),
	}

	for _, check := range checker.checks {
		status := "ok"
		if err := checkTCPDependency(ctx, check); err != nil {
			status = fmt.Sprintf("unreachable: %s", truncateErrorMessage(err.Error()))
			result.Status = "degraded"
		}
		result.Dependencies[check.name] = status
	}

	nominatimStatus := "ok"
	if err := checker.nominatim.Check(ctx); err != nil {
		nominatimStatus = fmt.Sprintf("unhealthy: %s", truncateErrorMessage(err.Error()))
		result.Status = "degraded"
	}
	result.Dependencies["nominatim"] = nominatimStatus

	return result
}

// Check returns an error if any dependency is unhealthy (for backwards compatibility with ReadinessChecker interface).
func (checker networkReadinessChecker) Check(ctx context.Context) error {
	result := checker.CheckAll(ctx)
	if result.Status != "ok" {
		for name, status := range result.Dependencies {
			if status != "ok" {
				return fmt.Errorf("dependency %s: %s", name, status)
			}
		}
	}
	return nil
}

func truncateErrorMessage(msg string) string {
	if len(msg) > 100 {
		return msg[:100]
	}
	return msg
}

func checkTCPDependency(ctx context.Context, check dependencyCheck) error {
	dialer := net.Dialer{Timeout: check.timeout}

	connection, err := dialer.DialContext(ctx, "tcp", check.address)
	if err != nil {
		return err
	}

	_ = connection.Close()

	return nil
}

func (check nominatimReadinessCheck) Check(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, check.statusURL, nil)
	if err != nil {
		return fmt.Errorf("build nominatim readiness request: %w", err)
	}

	client := http.Client{
		Timeout: check.timeout,
	}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("query nominatim status endpoint: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("nominatim status endpoint returned %d", response.StatusCode)
	}

	var status nominatimStatusResponse
	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		return fmt.Errorf("decode nominatim status response: %w", err)
	}

	if status.Status != 0 {
		return fmt.Errorf("nominatim status endpoint reported %d: %s", status.Status, status.Message)
	}

	return nil
}

func nominatimStatusURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	statusURL := *parsedURL
	statusURL.Path = path.Join(parsedURL.Path, "status")
	if statusURL.Path == "status" {
		statusURL.Path = "/status"
	}

	query := statusURL.Query()
	query.Set("format", "json")
	statusURL.RawQuery = query.Encode()

	return statusURL.String(), nil
}
