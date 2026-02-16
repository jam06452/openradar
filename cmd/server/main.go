package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"openradar/internal/config"
)

func main() {
	// cancel call
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// configss
	cfg := config.Load()
	print(cfg, ctx)

	// db
	// database, err := db.New(cfg.DatabaseURL)
	// if err != nil {
	// 	log.Fatalf("database init failed : %v", err)
	// }
}
