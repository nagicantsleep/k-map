package api

import (
	"context"
	"encoding/json"
	"net/http"
)

type statusResponse struct {
	Status       string            `json:"status"`
	Service      string            `json:"service,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

// ReadinessChecker verifies whether configured runtime dependencies are reachable.
type ReadinessChecker interface {
	Check(ctx context.Context) error
	CheckAll(ctx context.Context) ReadinessResult
}

func healthHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	writeStatusJSON(writer, http.StatusOK, statusResponse{
		Status:  "ok",
		Service: "k-map",
	})
}

func readinessHandler(checker ReadinessChecker) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		if checker == nil {
			writeStatusJSON(writer, http.StatusOK, statusResponse{
				Status:  "ok",
				Service: "k-map",
			})
			return
		}

		result := checker.CheckAll(request.Context())
		statusCode := http.StatusOK
		if result.Status != "ok" {
			statusCode = http.StatusServiceUnavailable
		}

		writeStatusJSON(writer, statusCode, statusResponse{
			Status:       result.Status,
			Dependencies: result.Dependencies,
		})
	}
}

func writeStatusJSON(writer http.ResponseWriter, statusCode int, response statusResponse) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	_ = json.NewEncoder(writer).Encode(response)
}
