package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	scanner.ScanJob()
	worker.Start(cfg, database)
	worker.Start(cfg, database)
	worker.Start(cfg, database)
	worker.Start(cfg, database)
	worker.Start(cfg, database)
	worker.Start(cfg, database)
	worker.Start(cfg, database)
	worker.Start(cfg, database)

	// When shutting down
	<-ctx.Done()

	log.Println("Shutting down server...")
}
