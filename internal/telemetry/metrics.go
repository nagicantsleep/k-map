package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metric collectors for the k-map API service.
type Metrics struct {
	RequestTotal          *prometheus.CounterVec
	RequestDuration       *prometheus.HistogramVec
	RateLimitRejections   *prometheus.CounterVec
	GeocoderDuration      *prometheus.HistogramVec
	CacheHits             *prometheus.CounterVec
	CacheMisses           *prometheus.CounterVec
}

// NewMetrics registers and returns all application metrics with the given registerer.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)

	return &Metrics{
		RequestTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "kmap_requests_total",
			Help: "Total number of HTTP requests processed, partitioned by endpoint, method, and status code.",
		}, []string{"endpoint", "method", "status_code"}),

		RequestDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "kmap_request_duration_seconds",
			Help:    "HTTP request latency in seconds, partitioned by endpoint and method.",
			Buckets: prometheus.DefBuckets,
		}, []string{"endpoint", "method"}),

		RateLimitRejections: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "kmap_rate_limit_rejections_total",
			Help: "Total number of requests rejected due to rate limiting, partitioned by endpoint.",
		}, []string{"endpoint"}),

		GeocoderDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "kmap_geocoder_request_duration_seconds",
			Help:    "Nominatim upstream call latency in seconds, partitioned by operation.",
			Buckets: prometheus.DefBuckets,
		}, []string{"operation"}),

		CacheHits: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "kmap_cache_hits_total",
			Help: "Total number of geocode cache hits, partitioned by operation.",
		}, []string{"operation"}),

		CacheMisses: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "kmap_cache_misses_total",
			Help: "Total number of geocode cache misses, partitioned by operation.",
		}, []string{"operation"}),
	}
}
