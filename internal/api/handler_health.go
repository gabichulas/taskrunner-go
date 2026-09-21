// Package api provides HTTP transport adapters and handlers for job ingestion and querying.
package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
