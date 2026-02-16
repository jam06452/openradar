package main

import (
	"context"
	"fmt"
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

	scanner.ScanJob()
	worker.Start()

	database, err := db.New(cfg.Database.URL)
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}
	fmt.Println(database) // Placeholder for now

	<-ctx.Done()

	log.Println("Shutting down server...")
}
