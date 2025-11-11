package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"digitalUniversity/application"
	"digitalUniversity/config"
	"digitalUniversity/logger"
)

func main() {
	logger := logger.GetInstance()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("config load failed: %v", err)
	}

	if err := logger.Initialize(cfg.LogDir, cfg.LogLevel); err != nil {
		logger.Fatalf("logger initialization failed: %v", err)
	}

	logger.Infof("Application starting. LogLevel=%d", cfg.LogLevel)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app := application.NewApplication()
	if err := app.Configure(cfg, logger, ctx); err != nil {
		logger.Fatalf("failed to configure application: %v", err)
	}

	logger.Info("bot has been successfully created and configured")

	app.Run(ctx)

	<-ctx.Done()

	logger.Info("Shutting down...")

	if err := app.DB.Close(); err != nil {
		logger.Errorf("failed to close DB: %v", err)
	}

	logger.Info("Application stopped")
}
