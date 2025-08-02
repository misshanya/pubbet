package main

import (
	"context"
	"github.com/misshanya/pubbet/internal/app"
	"github.com/misshanya/pubbet/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := setupLogger()

	cfg, err := config.New()
	if err != nil {
		logger.Error("failed to read config", slog.Any("error", err))
		os.Exit(1)
	}

	// Create app
	a, err := app.New(cfg, logger)
	if err != nil {
		logger.Error("failed to create app", slog.Any("error", err))
		os.Exit(1)
	}

	// Create ctx for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server
	errChan := make(chan error)
	go a.Start(errChan)

	// Read from channels
	// Exit with error OR gracefully shut down
	select {
	case err := <-errChan:
		logger.Error("failed to start server", slog.Any("error", err))
		os.Exit(1)
	case <-ctx.Done():
		a.Stop()
	}
}

func setupLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})

	logger := slog.New(handler)
	return logger
}
