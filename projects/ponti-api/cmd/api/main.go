package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
)

func main() {
	// Create a context with cancellation to handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture system signals for clean termination
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		<-sigChan // Wait for a signal
		log.Println("Received termination signal. Shutting down the application...")
		cancel()
	}()

	// Initialize dependencies using Wire
	deps, err := wire.Initialize()
	if err != nil {
		log.Fatalf("Error initializing dependencies: %s", err)
	}

	if err := RunGormMigrations(ctx, deps.GormRepository); err != nil {
		log.Fatalf("Failed to run Gorm's migrations: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := RunHttpServer(ctx, deps); err != nil {
			log.Fatalf("Error running HTTP server: %v", err)
		}
	}()

	wg.Wait()

	log.Println("Application terminated successfully.")
}
