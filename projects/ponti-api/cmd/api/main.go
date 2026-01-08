package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	wire "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/wire"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		<-sigChan
		log.Println("Received termination signal. Shutting down the application...")
		cancel()
	}()

	deps, err := wire.Initialize()
	if err != nil {
		log.Fatalf("Error initializing dependencies: %s", err)
	}

	setDeployEnv(ctx, deps)

	// Run the HTTP server
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := runHttpServer(ctx, deps); err != nil {
			log.Fatalf("Error running HTTP server: %v", err)
		}
	}()

	wg.Wait()

	log.Println("Application terminated successfully.")
}
