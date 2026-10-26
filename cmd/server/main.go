package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/ruslanty/url-checker/internal/checker"
	"gitlab.com/ruslanty/url-checker/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := checker.New(checker.ConfigFromEnv(), http.DefaultClient)
	server := &http.Server{
		Addr:              env("HTTP_ADDR", ":8123"),
		Handler:           httpapi.New(service),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	// goroutine keeps the HTTP API running while main waits for a signal
	go func() {
		logger.Info("server started", "addr", server.Addr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped", "error", err)

			os.Exit(1)
		}
	}()

	// channel receives the shutdown signal from operating system
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	// context and timeout give active requests time to finish gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
