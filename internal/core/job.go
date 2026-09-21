// Package core defines domain models, states, and secondary port interfaces for taskrunner-go.
package core

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type State string

const (
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateCompleted State = "completed"
	StateFailed    State = "failed"
)

type Job struct {
	ID         bson.ObjectID `bson:"_id,omitempty"`
	Task       string        `bson:"task_type"`
	State      State         `bson:"state"`
	Payload    any           `bson:"payload"`
	Result     any           `bson:"result,omitempty"`
	Error      string        `bson:"err,omitempty"`
	Attempts   int           `bson:"attempts"`
	MaxRetries int           `bson:"max_retries"`
	LockedAt   *time.Time    `bson:"locked_at,omitempty"`
	CreatedAt  time.Time     `bson:"created_at"`
	UpdatedAt  time.Time     `bson:"updated_at"`
	FinishedAt *time.Time    `bson:"finished_at,omitempty"`
}

type Repository interface {
	Enqueue(ctx context.Context, job *Job) (bson.ObjectID, error)
	ClaimJob(ctx context.Context) (*Job, error)
	CompleteJob(ctx context.Context, id bson.ObjectID, result any) error
	FailJob(ctx context.Context, id bson.ObjectID, errMsg string) error
	GetJobByID(ctx context.Context, id bson.ObjectID) (*Job, error)
}
