package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"openradar/internal/config"
	"openradar/internal/db"
	"openradar/internal/jobs"
	"openradar/internal/queue"
	"openradar/internal/server"
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

	hub := server.StartServer(database) // websocket

	for i := 0; i < cfg.Scanner.MaxConcurrentClones; i++ {
		worker.Start(ctx, cfg, database, hub)
	}

	jobContext := jobs.JobContext{
		DB:  database,
		Cfg: cfg,
		Ctx: ctx,
	}

	jobs.RunJobs(jobContext)

	// When shutting down
	<-ctx.Done()

	log.Println("Shutting down server...")
}
