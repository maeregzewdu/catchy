package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/maeregzewdu/catchy/internal/api"
	"github.com/maeregzewdu/catchy/internal/config"
	"github.com/maeregzewdu/catchy/internal/store"
)

// version is overridden at build time: go build -ldflags="-X main.version=1.0.0"
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: catchy <command> [flags]")
		fmt.Fprintln(os.Stderr, "commands:")
		fmt.Fprintln(os.Stderr, "  serve   start the catchy server")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		if err := runServe(os.Args[2:]); err != nil {
			slog.Error("fatal", "err", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	configPath := fs.String("config", "", "path to config.toml (default: ~/.catchy/config.toml)")
	logLevel := fs.String("log-level", "info", "log level: debug|info|warn|error")
	if err := fs.Parse(args); err != nil {
		return err
	}

	setupLogging(*logLevel)

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := os.MkdirAll(cfg.Data.Dir, 0755); err != nil {
		return fmt.Errorf("creating data dir %s: %w", cfg.Data.Dir, err)
	}

	catchyDir := config.DefaultDir()
	key, err := config.LoadOrCreateSecretKey(catchyDir)
	if err != nil {
		return fmt.Errorf("secret key: %w", err)
	}
	_ = key // used in Phase 3 for credential encryption

	db, err := store.New(cfg.Data.Dir)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	router := api.NewRouter(db, cfg, version)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("catchy started",
			"addr", addr,
			"trap", fmt.Sprintf("%s:%d", cfg.Trap.Host, cfg.Trap.Port),
			"version", version,
		)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server: %w", err)
	case sig := <-quit:
		slog.Info("shutting down", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	slog.Info("catchy stopped")
	return nil
}

func setupLogging(level string) {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l})))
}
