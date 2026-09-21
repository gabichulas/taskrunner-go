package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gabichulas/taskrunner-go/internal/api"
	"github.com/gabichulas/taskrunner-go/internal/store"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatalf("api server fatally ended: %v", err)
	}
}

func run(ctx context.Context) error {
	_ = godotenv.Load(".env")

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return errors.New("MONGODB_URI is not defined in environment or .env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to initialize mongo client: %w", err)
	}

	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		log.Println("[api] Disconnecting from MongoDB...")
		if err := client.Disconnect(disconnectCtx); err != nil {
			log.Printf("[api] Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Ping database
	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()
	if err := client.Ping(pingCtx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	log.Println("[api] Connected to MongoDB successfully")

	db := client.Database("taskrunner")
	repo := store.NewMongoStore(db.Collection("jobs"))

	apiServer := api.NewServer(repo)
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: apiServer.Routes(),
	}

	// Server runner in background goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("[api] HTTP server listening on port %s", port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server failure: %w", err)
	case <-ctx.Done():
		log.Println("[api] Shutting down HTTP server...")
	}

	// Graceful shutdown with 10s deadline
	shutdownCtx, cancelShutdown := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	log.Println("[api] HTTP server stopped cleanly")
	return nil
}
