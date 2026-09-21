// Package api provides HTTP transport adapters and handlers for job ingestion and querying.
package api

import (
	"net/http"

	"github.com/gabichulas/taskrunner-go/internal/core"
)

type Server struct {
	repo core.Repository
}

func NewServer(repo core.Repository) *Server {
	return &Server{
		repo: repo,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealthCheck)
	mux.HandleFunc("POST /jobs", s.handleEnqueueJob)
	mux.HandleFunc("GET /jobs/{id}", s.handleGetJobByID)
	mux.HandleFunc("GET /jobs", s.handleGetJobs)

	return mux
}
