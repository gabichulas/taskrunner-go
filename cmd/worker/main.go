package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gabichulas/taskrunner-go/internal/core"
	"github.com/gabichulas/taskrunner-go/internal/handler"
	"github.com/gabichulas/taskrunner-go/internal/store"
	"github.com/gabichulas/taskrunner-go/internal/worker"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatalf("taskrunner fatally ended: %v", err)
	}
}

func run(ctx context.Context) error {
	_ = godotenv.Load(".env")

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return errors.New("MONGODB_URI is not defined in environment or .env")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to initialize mongo client: %w", err)
	}

	// Guaranteed graceful disconnect when run() returns
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		log.Println("[worker] Disconnecting from MongoDB...")
		if err := client.Disconnect(disconnectCtx); err != nil {
			log.Printf("[worker] Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Fail fast: verify connection to MongoDB Atlas
	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()
	if err := client.Ping(pingCtx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB Atlas: %w", err)
	}
	log.Println("[worker] Connected to MongoDB Atlas successfully")

	db := client.Database("taskrunner")
	repo := store.NewMongoStore(db.Collection("jobs"))

	// Ensure compound index for atomic and fast ClaimJob operations
	if err := repo.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("failed to ensure indexes: %w", err)
	}
	log.Println("[worker] Compound indexes ensured ({ state: 1, created_at: 1 })")

	// Initialize worker pool and register handlers
	concurrency := 3
	pollInterval := 500 * time.Millisecond
	pool := worker.NewPool(repo, concurrency, pollInterval)

	if err := pool.Register("email", handler.Email); err != nil {
		return fmt.Errorf("failed to register handler: %w", err)
	}

	// Enqueue 3 sample jobs to demonstrate concurrent execution
	for i := 1; i <= 3; i++ {
		jobID, err := repo.Enqueue(ctx, &core.Job{
			Task:    "email",
			Payload: map[string]any{"recipient": fmt.Sprintf("test%d@example.com", i)},
		})
		if err == nil {
			log.Printf("[worker] Enqueued seed job ID: %s", jobID.Hex())
		}
	}

	log.Printf("[worker] Starting pool with %d workers (poll interval: %v)... Press Ctrl+C to stop", concurrency, pollInterval)
	pool.Start(ctx)
	log.Println("[worker] All workers shut down cleanly")

	return nil
}
