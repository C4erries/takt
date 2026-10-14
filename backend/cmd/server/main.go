package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"takt/internal/app"
	"takt/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

