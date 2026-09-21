// Package worker implements the concurrency orchestration engine and worker pool lifecycle.
package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gabichulas/taskrunner-go/internal/core"
)

type Handler func(ctx context.Context, job *core.Job) (any, error)

type Pool struct {
	repository   core.Repository
	handlers     map[string]Handler
	concurrency  int
	pollInterval time.Duration
}

func NewPool(repository core.Repository, concurrency int, pollInterval time.Duration) *Pool {
	return &Pool{
		repository:   repository,
		handlers:     make(map[string]Handler),
		concurrency:  concurrency,
		pollInterval: pollInterval,
	}
}

func (p *Pool) Register(task string, h Handler) error {
	if task == "" {
		return errors.New("task name cannot be empty")
	}
	if h == nil {
		return errors.New("handler cannot be nil")
	}
	if _, exists := p.handlers[task]; exists {
		return fmt.Errorf("handler already registered for task: %s", task)
	}
	p.handlers[task] = h
	return nil
}

// TODO: Option 3 - Implement Reaper/Sweeper goroutine to recover stale jobs where state is StateRunning and locked_at exceeds timeout.
func (p *Pool) Start(ctx context.Context) {
	var wg sync.WaitGroup

	for i := 1; i <= p.concurrency; i++ {
		wg.Add(1)
		go func(WorkerID int) {
			defer wg.Done()
			p.runWorker(ctx, WorkerID)
		}(i)
	}
	wg.Wait()
}

func (p *Pool) saveContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(parent), 5*time.Second)
}

func (p *Pool) runWorker(ctx context.Context, id int) {
	for {
		if ctx.Err() != nil {
			return
		}
		job, err := p.repository.ClaimJob(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(p.pollInterval):
				continue
			}
		}
		handler, ok := p.handlers[job.Task]
		if !ok {
			saveCtx, cancel := p.saveContext(ctx)
			p.repository.FailJob(saveCtx, job.ID, fmt.Sprintf("unregistered task handler: %s", job.Task))
			cancel()
			continue
		}
		res, err := handler(ctx, job)
		if err != nil {
			saveCtx, cancel := p.saveContext(ctx)
			p.repository.FailJob(saveCtx, job.ID, err.Error())
			cancel()
			continue
		}

		saveCtx, cancel := p.saveContext(ctx)
		p.repository.CompleteJob(saveCtx, job.ID, res)
		cancel()
	}
}
