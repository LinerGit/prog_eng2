package main

import (
	"context"
	"os/signal"
	"syscall"

	"ticket/internal/app"
	"ticket/internal/config"
	"ticket/internal/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg)
	application := app.New(log, cfg)

	if err := application.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("application stopped")
	}
}
