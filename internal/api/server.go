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

	return mux
}
