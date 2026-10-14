package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"takt/internal/infrastructure/config"
	"takt/internal/infrastructure/container"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := container.Run(ctx, cfg); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
