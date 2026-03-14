package telemetry

// Log field names used throughout the application.
// Use these constants to avoid typos and ensure consistency.
const (
	FieldRequestID       = "request_id"
	FieldTenantID        = "tenant_id"
	FieldMethod          = "method"
	FieldPath            = "path"
	FieldStatus          = "status"
	FieldLatencyMs       = "latency_ms"
	FieldGeocoderLatency = "geocoder_latency_ms"
	FieldError           = "error"
)
