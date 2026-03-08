package api

import (
	"context"
	"encoding/json"
	"net/http"
)

type statusResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// ReadinessChecker verifies whether configured runtime dependencies are reachable.
type ReadinessChecker interface {
	Check(ctx context.Context) error
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

		if checker != nil {
			if err := checker.Check(request.Context()); err != nil {
				writeStatusJSON(writer, http.StatusServiceUnavailable, statusResponse{
					Status:  "not_ready",
					Service: "k-map",
				})

				return
			}
		}

		writeStatusJSON(writer, http.StatusOK, statusResponse{
			Status:  "ready",
			Service: "k-map",
		})
	}
}

func writeStatusJSON(writer http.ResponseWriter, statusCode int, response statusResponse) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	_ = json.NewEncoder(writer).Encode(response)
}
