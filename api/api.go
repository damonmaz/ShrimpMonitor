package api

import (
	"encoding/json"
	"net/http"
)

// StartAPICPU serves the latest CPU values as JSON from http://localhost:8080/api/cpu.
func StartAPI(cpuSnapshot func() any, memSnapshot func() any) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/cpu", func(writer http.ResponseWriter, request *http.Request) {
		// Only allow GET requests for this endpoint
		if request.Method != http.MethodGet {
			writer.Header().Set("Allow", http.MethodGet)
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set the Content-Type header to application/json and encode the CPU snapshot as JSON
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(cpuSnapshot()); err != nil {
			http.Error(writer, "failed to encode CPU values", http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/api/memory", func(writer http.ResponseWriter, request *http.Request) {
		// Only allow GET requests for this endpoint
		if request.Method != http.MethodGet {
			writer.Header().Set("Allow", http.MethodGet)
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set the Content-Type header to application/json and encode the Memory snapshot as JSON
		writer.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(writer).Encode(memSnapshot()); err != nil {
			http.Error(writer, "failed to encode Memory values", http.StatusInternalServerError)
		}
	})
	return http.ListenAndServe("127.0.0.1:8080", mux)
}

// func StartAPIMemory(memSnapshot func() any) error {
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/api/memory", func(writer http.ResponseWriter, request *http.Request) {
// 		// Only allow GET requests for this endpoint
// 		if request.Method != http.MethodGet {
// 			writer.Header().Set("Allow", http.MethodGet)
// 			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		// Set the Content-Type header to application/json and encode the Memory snapshot as JSON
// 		writer.Header().Set("Content-Type", "application/json")
// 		if err := json.NewEncoder(writer).Encode(memSnapshot()); err != nil {
// 			http.Error(writer, "failed to encode Memory values", http.StatusInternalServerError)
// 		}
// 	})
// 	return http.ListenAndServe("127.0.0.1:8080", mux)
// }
