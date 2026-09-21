// Package handler implements concrete job task execution strategies.
package handler

import (
	"context"
	"log"
	"time"

	"github.com/gabichulas/taskrunner-go/internal/core"
)

// Email processes email jobs, respecting context cancellation.
func Email(ctx context.Context, job *core.Job) (any, error) {
	log.Printf("[handler:email] Sending email for job ID: %s, payload: %v", job.ID.Hex(), job.Payload)

	select {
	case <-time.After(1 * time.Second):
		log.Printf("[handler:email] Email sent successfully for job ID: %s", job.ID.Hex())
		return map[string]any{
			"status":       "delivered",
			"delivered_at": time.Now().UTC(),
		}, nil
	case <-ctx.Done():
		log.Printf("[handler:email] Email canceled for job ID: %s", job.ID.Hex())
		return nil, ctx.Err()
	}
}
