package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	httpadapter "github.com/memlore/memlore/internal/adapters/http"
	mcpadapter "github.com/memlore/memlore/internal/adapters/mcp"
	"github.com/memlore/memlore/internal/bootstrap"
	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/graphclient"
	"github.com/memlore/memlore/internal/infrastructure/postgres"
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
	case "migrate":
		if err := runMigrate(logger); err != nil {
			fmt.Fprintf(os.Stderr, "migrate failed: %v\n", err)
			return 1
		}
		return 0
	case "worker":
		if err := runWorker(logger); err != nil {
			fmt.Fprintf(os.Stderr, "worker failed: %v\n", err)
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
	fmt.Fprintln(stdout, "  memlore migrate")
	fmt.Fprintln(stdout, "  memlore serve")
	fmt.Fprintln(stdout, "  memlore mcp")
	fmt.Fprintln(stdout, "  memlore worker")
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
	graphURL := envOr("MEMLORE_GRAPH_SERVICE_URL", "http://127.0.0.1:8090")
	graph := graphclient.NewClient(graphURL, nil)
	handler := httpadapter.NewHandlers(begin, clock.SystemClock{}, graph, Version).Router()
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

func runMigrate(logger *slog.Logger) error {
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	if err := bootstrap.RunMigrations(dsn); err != nil {
		return err
	}
	logger.Info("memlore migrate complete", "dsn", redactDSN(dsn))
	return nil
}

func runWorker(logger *slog.Logger) error {
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	graphURL := envOr("MEMLORE_GRAPH_SERVICE_URL", "http://127.0.0.1:8090")
	graph := graphclient.NewClient(graphURL, nil)
	runner := postgres.NewOutboxProcessor(pool)
	handler := commands.NewProcessOutboxHandler(runner, graph, clock.SystemClock{}, 10)

	interval := workerPollInterval()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("memlore worker started", "graph_service", graphURL, "poll_interval", interval.String())
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		processed, err := handler.ProcessOnce(ctx)
		if err != nil {
			logger.Error("outbox processing failed", "error", err)
		} else if processed > 0 {
			logger.Info("outbox batch processed", "count", processed)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func workerPollInterval() time.Duration {
	raw := envOr("MEMLORE_WORKER_POLL_INTERVAL", "5s")
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 5 * time.Second
	}
	return d
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
	graphURL := envOr("MEMLORE_GRAPH_SERVICE_URL", "http://127.0.0.1:8090")
	graph := graphclient.NewClient(graphURL, nil)
	server := mcpadapter.NewServer(begin, clock.SystemClock{}, graph, Version, logger)
	logger.Info("memlore mcp listening on stdio")
	return server.Run(context.Background(), &sdkmcp.StdioTransport{})
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func redactDSN(dsn string) string {
	// Avoid logging passwords; host/db tail is enough for ops.
	if idx := strings.LastIndex(dsn, "@"); idx >= 0 && idx+1 < len(dsn) {
		return dsn[idx+1:]
	}
	return "configured"
}
