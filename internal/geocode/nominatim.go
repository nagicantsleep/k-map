package geocode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/nagicantsleep/k-map/internal/api"
	"github.com/nagicantsleep/k-map/internal/telemetry"
)

// NominatimClient implements the api.Geocoder interface using Nominatim.
type NominatimClient struct {
	baseURL    string
	httpClient *http.Client
	metrics    *telemetry.Metrics
	maxRetries int
}

// NewNominatimClient creates a new Nominatim client.
func NewNominatimClient(baseURL string, timeout time.Duration) *NominatimClient {
	return &NominatimClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: 1,
	}
}

// WithRetries configures the maximum number of retries for transient failures.
func (c *NominatimClient) WithRetries(n int) *NominatimClient {
	c.maxRetries = n
	return c
}

// WithMetrics attaches a metrics collector to the Nominatim client.
func (c *NominatimClient) WithMetrics(m *telemetry.Metrics) *NominatimClient {
	c.metrics = m
	return c
}

// nominatimResult represents a single Nominatim search result.
type nominatimResult struct {
	PlaceID     int                    `json:"place_id"`
	Licence     string                 `json:"licence"`
	OsmType     string                 `json:"osm_type"`
	OsmID       int64                  `json:"osm_id"`
	Lat         string                 `json:"lat"`
	Lon         string                 `json:"lon"`
	DisplayName string                 `json:"display_name"`
	Class       string                 `json:"class"`
	Type        string                 `json:"type"`
	Importance  float64                `json:"importance"`
	Address     *nominatimAddress      `json:"address"`
}

// nominatimAddress represents address components from Nominatim.
type nominatimAddress struct {
	HouseNumber  string `json:"house_number"`
	Road         string `json:"road"`
	City         string `json:"city"`
	Town         string `json:"town"`
	Village      string `json:"village"`
	State        string `json:"state"`
	Postcode     string `json:"postcode"`
	Country      string `json:"country"`
	CountryCode  string `json:"country_code"`
}

// Search performs a forward geocoding search using Nominatim.
func (c *NominatimClient) Search(ctx context.Context, query string, limit int) ([]api.GeocodeResult, error) {
	if limit <= 0 {
		limit = 10
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("limit", strconv.Itoa(limit))
	params.Set("addressdetails", "1")

	endpoint := fmt.Sprintf("%s/search?%s", c.baseURL, params.Encode())

	body, err := c.doWithRetry(ctx, endpoint, "search")
	if err != nil {
		return nil, err
	}

	var results []nominatimResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return c.normalizeResults(results), nil
}

// Reverse performs a reverse geocoding lookup using Nominatim.
func (c *NominatimClient) Reverse(ctx context.Context, lat, lng float64) (*api.GeocodeResult, error) {
	params := url.Values{}
	params.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	params.Set("lon", strconv.FormatFloat(lng, 'f', -1, 64))
	params.Set("format", "json")
	params.Set("addressdetails", "1")

	endpoint := fmt.Sprintf("%s/reverse?%s", c.baseURL, params.Encode())

	body, err := c.doWithRetry(ctx, endpoint, "reverse")
	if err != nil {
		return nil, err
	}

	var result nominatimResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Nominatim returns an empty result when no match is found
	if result.PlaceID == 0 && result.Lat == "" {
		return nil, nil
	}

	normalized := c.normalizeResult(result)
	return &normalized, nil
}

// doWithRetry executes a GET request against the given endpoint with exponential-backoff retry
// on transient failures (network errors or 5xx). It does not retry on 4xx or context cancellation.
func (c *NominatimClient) doWithRetry(ctx context.Context, endpoint, operation string) ([]byte, error) {
	const initialBackoff = 500 * time.Millisecond

	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := initialBackoff * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		body, transient, err := c.doOnce(ctx, endpoint, operation)
		if err == nil {
			return body, nil
		}

		lastErr = err
		if !transient {
			return nil, err
		}
	}

	return nil, lastErr
}

// doOnce performs a single HTTP GET. Returns (body, transient, error).
// transient=true means the caller may retry; transient=false means do not retry.
func (c *NominatimClient) doOnce(ctx context.Context, endpoint, operation string) ([]byte, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	elapsed := time.Since(start)

	if c.metrics != nil {
		c.metrics.GeocoderDuration.WithLabelValues(operation).Observe(elapsed.Seconds())
	}

	if err != nil {
		// Network-level errors (timeout, connection refused) are transient
		// unless the context was cancelled.
		if ctx.Err() != nil {
			return nil, false, fmt.Errorf("nominatim request cancelled: %w", ctx.Err())
		}
		return nil, true, fmt.Errorf("nominatim request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("nominatim returned status %d", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		// 4xx: non-transient, do not retry
		return nil, false, fmt.Errorf("nominatim returned status %d", resp.StatusCode)
	}

	return body, false, nil
}

func (c *NominatimClient) normalizeResults(results []nominatimResult) []api.GeocodeResult {
	if len(results) == 0 {
		return nil
	}

	normalized := make([]api.GeocodeResult, len(results))
	for i, r := range results {
		normalized[i] = c.normalizeResult(r)
	}

	return normalized
}

func (c *NominatimClient) normalizeResult(r nominatimResult) api.GeocodeResult {
	lat, _ := strconv.ParseFloat(r.Lat, 64)
	lng, _ := strconv.ParseFloat(r.Lon, 64)

	result := api.GeocodeResult{
		FormattedAddress: r.DisplayName,
		Latitude:         lat,
		Longitude:        lng,
		Confidence:       r.Importance,
		Source:           "osm",
		PlaceType:        r.Type,
	}

	if r.Address != nil {
		result.Components = api.AddressComponents{
			StreetNumber: r.Address.HouseNumber,
			Street:       r.Address.Road,
			State:        r.Address.State,
			PostalCode:   r.Address.Postcode,
			Country:      r.Address.Country,
			CountryCode:  r.Address.CountryCode,
		}

		// Handle city/town/village
		if r.Address.City != "" {
			result.Components.City = r.Address.City
		} else if r.Address.Town != "" {
			result.Components.City = r.Address.Town
		} else if r.Address.Village != "" {
			result.Components.City = r.Address.Village
		}
	}

	return result
}
