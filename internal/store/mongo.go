// Package store provides persistence adapters for taskrunner-go using MongoDB.
package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gabichulas/taskrunner-go/internal/core"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoStore struct {
	collection *mongo.Collection
}

func NewMongoStore(collection *mongo.Collection) *MongoStore {
	return &MongoStore{
		collection: collection,
	}
}

func (s *MongoStore) ClaimJob(ctx context.Context) (*core.Job, error) {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"state":      core.StateRunning,
			"locked_at":  now,
			"updated_at": now,
		},
		"$inc": bson.M{
			"attempts": 1,
		},
	}
	job := core.Job{}
	opts := options.FindOneAndUpdate().SetSort(bson.M{"created_at": 1}).SetReturnDocument(options.After)
	err := s.collection.FindOneAndUpdate(ctx, bson.M{"state": core.StateQueued}, update, opts).Decode(&job)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, core.ErrJobNotFound
		}
		return nil, err
	}

	return &job, nil
}

func (s *MongoStore) CompleteJob(ctx context.Context, id bson.ObjectID, result any) error {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"state":       core.StateCompleted,
			"result":      result,
			"updated_at":  now,
			"finished_at": now,
		},
	}

	res, err := s.collection.UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("job not found: %v", id)
	}
	return nil
}

func (s *MongoStore) FailJob(ctx context.Context, id bson.ObjectID, errMsg string) error {
	// TODO: Option 1 - Implement retry mechanism: if attempts < max_retries, transition state to StateQueued with backoff instead of StateFailed.
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"state":       core.StateFailed,
			"err":         errMsg,
			"updated_at":  now,
			"finished_at": now,
			"locked_at":   nil,
		},
	}

	res, err := s.collection.UpdateByID(ctx, id, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("job not found: %v", id)
	}
	return nil
}

func (s *MongoStore) GetJobByID(ctx context.Context, id bson.ObjectID) (*core.Job, error) {
	job := core.Job{}
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&job)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, core.ErrJobNotFound
		}
		return nil, err
	}

	return &job, nil
}

func (s *MongoStore) Enqueue(ctx context.Context, job *core.Job) (bson.ObjectID, error) {
	if job.ID.IsZero() {
		job.ID = bson.NewObjectID()
	}

	job.State = core.StateQueued
	job.Attempts = 0
	job.MaxRetries = 5
	job.CreatedAt = time.Now().UTC()
	job.UpdatedAt = time.Now().UTC()
	_, err := s.collection.InsertOne(ctx, job)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return bson.NilObjectID, core.ErrJobNotFound
		}
		return bson.NilObjectID, err
	}
	return job.ID, nil
}

func (s *MongoStore) EnsureIndexes(ctx context.Context) error {
	model := mongo.IndexModel{
		Keys: bson.D{
			{Key: "state", Value: 1},
			{Key: "created_at", Value: 1},
		},
	}
	_, err := s.collection.Indexes().CreateOne(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	return nil
}
