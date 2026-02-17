package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"openradar/internal/config"
	"openradar/internal/db"
	"openradar/internal/queue"
	"openradar/internal/scanner"
	"openradar/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	queue.NewInMemoryQueue(100)

	database, err := db.New(cfg.Database.URL)
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}

	for i := 0; i < cfg.Scanner.MaxConcurrentClones; i++ {
		worker.Start(cfg, database)
	}

	go func() {
		ticker := time.NewTicker(35 * time.Second) // Scan every 35 sec
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("scanning for latest repo updates")
				scanner.ScanJob(cfg.GitHub.Key)
			case <-ctx.Done():
				return
			}
		}
	}()

	// When shutting down
	<-ctx.Done()

	log.Println("Shutting down server...")
}
