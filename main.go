package main

import (
	"context"
	"go-bot/app"
	"go-bot/config"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() { rand.Seed(time.Now().UnixNano()) }

func main() {
	cfg := config.New("config.json")
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	app := app.New(cfg, logger)

	go func() {
		app.Run(ctx)
	}()

	sig := <-sigChan
	logger.Info("Received shutdown signal", "signal", sig)

	cancel()

	logger.Info("Shutting down gracefully...")
}
