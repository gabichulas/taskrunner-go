package worker

import (
	"context"
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

func NewPool(repository core.Repository, handlers map[string]Handler, concurrency int, pollInterval time.Duration) *Pool {
	pool := Pool{
		repository:   repository,
		handlers:     handlers,
		concurrency:  concurrency,
		pollInterval: pollInterval,
	}
	return &pool
}

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

func (p *Pool) runWorker(ctx context.Context, id int) {
	return
}
