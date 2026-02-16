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
)

func main() {
	// cancel call
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// configss
	cfg := config.Load()

	// db
	database, err := db.New(cfg.Database.URL)
	if err != nil {
		log.Fatalf("database init failed : %v", err)
	}

	fmt.Println(database)

	<-ctx.Done()
	fmt.Println("shutting down server")
}
