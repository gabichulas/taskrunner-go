// Package api provides HTTP transport adapters and handlers for job ingestion and querying.
package api

import "time"

type EnqueueJobRequest struct {
	Task    string `json:"task"`
	Payload any    `json:"payload"`
}

type JobResponse struct {
	ID         string     `json:"id"`
	Task       string     `json:"task"`
	State      string     `json:"state"`
	Payload    any        `json:"payload"`
	Result     any        `json:"result,omitempty"`
	Error      string     `json:"error,omitempty"`
	Attempts   int        `json:"attempts"`
	MaxRetries int        `json:"max_retries"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}
