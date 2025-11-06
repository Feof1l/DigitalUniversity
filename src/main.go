package main

import (
	"context"
	"os"
	"os/signal"

	_ "github.com/lib/pq"

	"digitalUniversity/application"
	"digitalUniversity/config"
	"digitalUniversity/logger"
)

func main() {
    logr := logger.GetInstance()

    cfg, err := config.Load()
    if err != nil {
        logr.Fatalf("config load failed: %v", err)
    }

    if err := logr.Initialize(cfg.LogDir, cfg.LogLevel); err != nil {
        logr.Fatalf("logger initialization failed: %v", err)
    }

    logr.Infof("Application starting. LogLevel=%d", cfg.LogLevel)

    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()

    app := application.NewApplication()
    app.Configure(cfg, ctx)
    app.Run(ctx)

    logr.Info("Application stopped")
}
