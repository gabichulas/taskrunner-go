// Package handler implements concrete job task execution strategies.
package handler

import (
	"github.com/gabichulas/taskrunner-go/internal/worker"
)

// NewRegistry returns the default mapping of task names to their respective handlers.
func NewRegistry() map[string]worker.Handler {
	return map[string]worker.Handler{
		"email": Email,
	}
}
