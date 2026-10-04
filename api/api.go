package api

import (
	"encoding/json"
	"net/http"
)

// Starts the API server and sets up the endpoints for monitoring.
func StartAPI(cpuSnapshot func() any, memSnapshot func() any) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/cpu", jsonEndpointGET(cpuSnapshot))
	mux.HandleFunc("/api/memory", jsonEndpointGET(memSnapshot))
	return http.ListenAndServe("127.0.0.1:8080", mux)
}

// Creates a GET handler that returns the supplied snapshot as JSON.
func jsonEndpointGET(snapshot func() any) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.Header().Set("Allow", http.MethodGet)
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(snapshot()); err != nil {
			http.Error(writer, "failed to encode response", http.StatusInternalServerError)
		}
	}
}
