// Package api provides HTTP transport adapters and handlers for job ingestion and querying.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/gabichulas/taskrunner-go/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *Server) handleEnqueueJob(w http.ResponseWriter, r *http.Request) {
	var enq EnqueueJobRequest

	err := json.NewDecoder(r.Body).Decode(&enq)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if enq.Task == "" {
		http.Error(w, "task cannot be empty", http.StatusBadRequest)
		return
	} else if enq.Payload == nil {
		http.Error(w, "payload cannot be empty", http.StatusBadRequest)
		return
	}

	id, err := s.repo.Enqueue(r.Context(), &core.Job{Task: enq.Task, Payload: enq.Payload})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id.Hex()})
}

func (s *Server) handleGetJobByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "id unparsable", http.StatusBadRequest)
		return
	}
	job, err := s.repo.GetJobByID(r.Context(), id)
	if err != nil && err == core.ErrJobNotFound {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	resp := JobResponse{
		ID:         id.Hex(),
		Task:       job.Task,
		Payload:    job.Payload,
		State:      string(job.State),
		Result:     job.Result,
		Error:      job.Error,
		Attempts:   job.Attempts,
		MaxRetries: job.MaxRetries,
		CreatedAt:  job.CreatedAt,
		UpdatedAt:  job.UpdatedAt,
		FinishedAt: job.FinishedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGetJobs(w http.ResponseWriter, r *http.Request) {
	return
}
