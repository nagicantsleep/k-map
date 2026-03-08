package api

import (
	"encoding/json"
	"net/http"
)

type statusResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
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

func readinessHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	writeStatusJSON(writer, http.StatusOK, statusResponse{
		Status:  "ready",
		Service: "k-map",
	})
}

func writeStatusJSON(writer http.ResponseWriter, statusCode int, response statusResponse) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	_ = json.NewEncoder(writer).Encode(response)
}
