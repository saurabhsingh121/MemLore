package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	httpadapter "github.com/memlore/memlore/internal/adapters/http"
	mcpadapter "github.com/memlore/memlore/internal/adapters/mcp"
	"github.com/memlore/memlore/internal/bootstrap"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Version is the MemLore Core build version.
const Version = "0.1.0-dev"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	os.Exit(run(os.Args, os.Stdout, logger))
}

func run(args []string, stdout io.Writer, logger *slog.Logger) int {
	if len(args) < 2 {
		printUsage(stdout)
		return 0
	}
	switch args[1] {
	case "version":
		fmt.Fprintln(stdout, Version)
		logger.Info("memlore version", "version", Version)
		return 0
	case "serve":
		host := envOr("MEMLORE_HTTP_HOST", "127.0.0.1")
		port := envOr("MEMLORE_HTTP_PORT", "8080")
		addr := fmt.Sprintf("%s:%s", host, port)
		if err := runServe(addr, logger); err != nil {
			fmt.Fprintf(os.Stderr, "serve failed: %v\n", err)
			return 1
		}
		return 0
	case "mcp":
		if err := runMCP(logger); err != nil {
			fmt.Fprintf(os.Stderr, "mcp failed: %v\n", err)
			return 1
		}
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[1])
		return 1
	}
}

func printUsage(stdout io.Writer) {
	fmt.Fprintf(stdout, "memlore %s — MemLore Core (Go)\n", Version)
	fmt.Fprintln(stdout, "usage:")
	fmt.Fprintln(stdout, "  memlore version")
	fmt.Fprintln(stdout, "  memlore serve")
	fmt.Fprintln(stdout, "  memlore mcp")
}

func runServe(addr string, logger *slog.Logger) error {
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	handler := httpadapter.NewHandlers(begin, clock.SystemClock{}, Version).Router()
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("memlore serve listening", "addr", "http://"+addr)
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func runMCP(logger *slog.Logger) error {
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	server := mcpadapter.NewServer(begin, clock.SystemClock{}, Version, logger)
	logger.Info("memlore mcp listening on stdio")
	return server.Run(context.Background(), &sdkmcp.StdioTransport{})
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
