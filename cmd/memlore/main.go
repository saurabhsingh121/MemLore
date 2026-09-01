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
	authadapter "github.com/memlore/memlore/internal/adapters/auth"
	"github.com/memlore/memlore/internal/adapters/cli"
	httpadapter "github.com/memlore/memlore/internal/adapters/http"
	mcpadapter "github.com/memlore/memlore/internal/adapters/mcp"
	"github.com/memlore/memlore/internal/adapters/presenters"
	appauth "github.com/memlore/memlore/internal/application/auth"
	"github.com/memlore/memlore/internal/application/authz"
	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/bootstrap"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/fsadr"
	"github.com/memlore/memlore/internal/infrastructure/gitcli"
	"github.com/memlore/memlore/internal/infrastructure/githubhttp"
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
	case "profile":
		opts, err := cli.ParseProfileArgs(args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "profile failed: %v\n", err)
			return 1
		}
		if err := runProfile(opts, stdout, logger); err != nil {
			fmt.Fprintf(os.Stderr, "profile failed: %v\n", err)
			return 1
		}
		return 0
	case "context":
		opts, err := cli.ParseContextArgs(args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "context failed: %v\n", err)
			return 1
		}
		if err := runContext(opts, stdout, logger); err != nil {
			fmt.Fprintf(os.Stderr, "context failed: %v\n", err)
			return 1
		}
		return 0
	case "migrate":
		if err := runMigrate(logger); err != nil {
			fmt.Fprintf(os.Stderr, "migrate failed: %v\n", err)
			return 1
		}
		return 0
	case "decision":
		if len(args) < 3 {
			printUsage(stdout)
			return 1
		}
		switch args[2] {
		case "create":
			opts, err := cli.ParseDecisionCreateArgs(args[3:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "decision create failed: %v\n", err)
				return 1
			}
			if opts.Actor == "" {
				opts.Actor = strings.TrimSpace(os.Getenv("MEMLORE_ACTOR"))
			}
			if err := runDecisionCreate(opts, stdout, logger); err != nil {
				fmt.Fprintf(os.Stderr, "decision create failed: %v\n", err)
				return 1
			}
			return 0
		case "get":
			opts, err := cli.ParseDecisionGetArgs(args[3:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "decision get failed: %v\n", err)
				return 1
			}
			if err := runDecisionGet(opts, stdout, logger); err != nil {
				fmt.Fprintf(os.Stderr, "decision get failed: %v\n", err)
				return 1
			}
			return 0
		case "list":
			opts, err := cli.ParseDecisionListArgs(args[3:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "decision list failed: %v\n", err)
				return 1
			}
			if err := runDecisionList(opts, stdout, logger); err != nil {
				fmt.Fprintf(os.Stderr, "decision list failed: %v\n", err)
				return 1
			}
			return 0
		case "supersede":
			opts, err := cli.ParseDecisionSupersedeArgs(args[3:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "decision supersede failed: %v\n", err)
				return 1
			}
			if opts.Actor == "" {
				opts.Actor = strings.TrimSpace(os.Getenv("MEMLORE_ACTOR"))
			}
			if err := runDecisionSupersede(opts, stdout, logger); err != nil {
				fmt.Fprintf(os.Stderr, "decision supersede failed: %v\n", err)
				return 1
			}
			return 0
		default:
			fmt.Fprintf(os.Stderr, "unknown decision command: %s\n", args[2])
			return 1
		}
	case "review":
		if len(args) < 3 {
			printUsage(stdout)
			return 1
		}
		switch args[2] {
		case "list":
			opts, err := cli.ParseReviewListArgs(args[3:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "review list failed: %v\n", err)
				return 1
			}
			if err := runReviewList(opts, stdout, logger); err != nil {
				fmt.Fprintf(os.Stderr, "review list failed: %v\n", err)
				return 1
			}
			return 0
		case "accept":
			opts, err := cli.ParseReviewAcceptArgs(args[3:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "review accept failed: %v\n", err)
				return 1
			}
			if opts.Actor == "" {
				opts.Actor = strings.TrimSpace(os.Getenv("MEMLORE_ACTOR"))
			}
			if err := runReviewAccept(opts, stdout, logger); err != nil {
				fmt.Fprintf(os.Stderr, "review accept failed: %v\n", err)
				return 1
			}
			return 0
		case "reject":
			opts, err := cli.ParseReviewRejectArgs(args[3:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "review reject failed: %v\n", err)
				return 1
			}
			if opts.Actor == "" {
				opts.Actor = strings.TrimSpace(os.Getenv("MEMLORE_ACTOR"))
			}
			if err := runReviewReject(opts, stdout, logger); err != nil {
				fmt.Fprintf(os.Stderr, "review reject failed: %v\n", err)
				return 1
			}
			return 0
		default:
			fmt.Fprintf(os.Stderr, "unknown review command: %s\n", args[2])
			return 1
		}
	case "ingest":
		if len(args) < 3 {
			printUsage(stdout)
			return 1
		}
		switch args[2] {
		case "git":
			opts, err := cli.ParseIngestGitArgs(args[3:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "ingest git failed: %v\n", err)
				return 1
			}
			if opts.Actor == "" {
				opts.Actor = strings.TrimSpace(os.Getenv("MEMLORE_ACTOR"))
			}
			if err := runIngestGit(opts, stdout, logger); err != nil {
				fmt.Fprintf(os.Stderr, "ingest git failed: %v\n", err)
				return 1
			}
			return 0
		case "pr":
			opts, err := cli.ParseIngestPRArgs(args[3:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "ingest pr failed: %v\n", err)
				return 1
			}
			if opts.Actor == "" {
				opts.Actor = strings.TrimSpace(os.Getenv("MEMLORE_ACTOR"))
			}
			if err := runIngestPR(opts, stdout, logger); err != nil {
				fmt.Fprintf(os.Stderr, "ingest pr failed: %v\n", err)
				return 1
			}
			return 0
		case "adr":
			opts, err := cli.ParseIngestADRArgs(args[3:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "ingest adr failed: %v\n", err)
				return 1
			}
			if opts.Actor == "" {
				opts.Actor = strings.TrimSpace(os.Getenv("MEMLORE_ACTOR"))
			}
			if err := runIngestADR(opts, stdout, logger); err != nil {
				fmt.Fprintf(os.Stderr, "ingest adr failed: %v\n", err)
				return 1
			}
			return 0
		case "status":
			opts, err := cli.ParseIngestStatusArgs(args[3:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "ingest status failed: %v\n", err)
				return 1
			}
			if err := runIngestStatus(opts, stdout, logger); err != nil {
				fmt.Fprintf(os.Stderr, "ingest status failed: %v\n", err)
				return 1
			}
			return 0
		default:
			fmt.Fprintf(os.Stderr, "unknown ingest command: %s\n", args[2])
			return 1
		}
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
	fmt.Fprintln(stdout, "  memlore profile --repository <key>")
	fmt.Fprintln(stdout, "  memlore context --task <text> --repository <key>")
	fmt.Fprintln(stdout, "  memlore ingest git --repository <key> --path <dir> [--actor <id>]")
	fmt.Fprintln(stdout, "  memlore ingest pr --repository <key> [--pr <n>] [--actor <id>]")
	fmt.Fprintln(stdout, "  memlore ingest adr --repository <key> --path <dir> [--adr-dir <rel>] [--actor <id>]")
	fmt.Fprintln(stdout, "  memlore ingest status --repository <key> [--kind git|pr|adr]")
	fmt.Fprintln(stdout, "  memlore review list --repository <key>")
	fmt.Fprintln(stdout, "  memlore review accept <id> [--statement <text>] [--actor <id>]")
	fmt.Fprintln(stdout, "  memlore review reject <id> [--actor <id>]")
	fmt.Fprintln(stdout, "  memlore decision create --repository <key> --question <text> --choice <text> --owner <id> [--actor <id>]")
	fmt.Fprintln(stdout, "  memlore decision get <id>")
	fmt.Fprintln(stdout, "  memlore decision list --repository <key>")
	fmt.Fprintln(stdout, "  memlore decision supersede <id> --question <text> --choice <text> --owner <id> [--actor <id>]")
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
	handlers := httpadapter.NewHandlers(begin, clock.SystemClock{}, graph, Version)
	authSvc, err := newAuthService()
	if err != nil {
		return err
	}
	handlers.Auth = authSvc
	handlers.Membership = postgres.NewMembershipDirectory(pool)
	handlers.Authz = &authz.Gate{Auth: authSvc, Membership: handlers.Membership}
	handler := handlers.Router()
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
	tools := mcpadapter.NewTools(begin, clock.SystemClock{}, graph)
	authSvc, err := newAuthService()
	if err != nil {
		return err
	}
	tools.Auth = authSvc
	tools.Membership = postgres.NewMembershipDirectory(pool)
	tools.Authz = &authz.Gate{Auth: authSvc, Membership: tools.Membership}
	server := mcpadapter.NewServerFromTools(tools, Version, logger)
	logger.Info("memlore mcp listening on stdio", "oidc", authSvc.Config.Enabled())
	return server.Run(context.Background(), &sdkmcp.StdioTransport{})
}

func runProfile(opts cli.ProfileArgs, stdout io.Writer, logger *slog.Logger) error {
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
	list := queries.NewListLoreByScopeHandler(begin)
	search := queries.NewSearchKnowledgeHandler(begin, graph, nil)
	handler := queries.NewRepositoryProfileHandler(list, search)
	scope, err := domain.NewScope(domain.ScopeKindRepository, opts.Repository)
	if err != nil {
		return err
	}
	result, err := handler.Handle(context.Background(), queries.RepositoryProfileQuery{
		Scope:       scope,
		TokenBudget: opts.TokenBudget,
	})
	if err != nil {
		return err
	}
	_, err = io.WriteString(stdout, cli.FormatProfile(presenters.ToRepositoryProfile(result)))
	return err
}

func runContext(opts cli.ContextArgs, stdout io.Writer, logger *slog.Logger) error {
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
	list := queries.NewListLoreByScopeHandler(begin)
	search := queries.NewSearchKnowledgeHandler(begin, graph, nil)
	handler := queries.NewCompileContextHandler(search, list)
	handler.SetDecisions(queries.NewListDecisionsHandler(begin))
	scope, err := domain.NewScope(domain.ScopeKindRepository, opts.Repository)
	if err != nil {
		return err
	}
	result, err := handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:         opts.Task,
		Query:        opts.Query,
		Scope:        scope,
		TokenBudget:  opts.TokenBudget,
		Branch:       opts.Branch,
		Ticket:       opts.Ticket,
		ChangedFiles: opts.ChangedFiles,
		WorkingFiles: opts.WorkingFiles,
		AgentID:      opts.AgentID,
	})
	if err != nil {
		return err
	}
	_, err = io.WriteString(stdout, cli.FormatContext(presenters.ToContextPacket(result)))
	return err
}

func runIngestGit(opts cli.IngestGitArgs, stdout io.Writer, logger *slog.Logger) error {
	if strings.TrimSpace(opts.Actor) == "" {
		return fmt.Errorf("validation_error: --actor or MEMLORE_ACTOR is required")
	}
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	scope, err := domain.NewScope(domain.ScopeKindRepository, opts.Repository)
	if err != nil {
		return err
	}
	handler := commands.NewIngestGitHandler(begin, clock.SystemClock{}, gitcli.NewReader())
	logger.Info("memlore ingest git started", "repository_id", opts.Repository, "path", opts.Path)
	run, err := handler.Handle(context.Background(), commands.IngestGitCommand{
		Scope:      scope,
		Path:       opts.Path,
		ActorID:    opts.Actor,
		MaxCommits: opts.MaxCommits,
	})
	if err != nil {
		logger.Error("memlore ingest git failed", "repository_id", opts.Repository, "error", err)
		return err
	}
	if run.Status == domain.IngestRunFailed {
		logger.Error("memlore ingest git failed", "repository_id", opts.Repository, "run_id", run.ID, "error", run.ErrorSummary)
	} else {
		logger.Info("memlore ingest git completed",
			"repository_id", opts.Repository,
			"run_id", run.ID,
			"commits_seen", run.CommitsSeen,
			"candidates_stored", run.CandidatesStored,
		)
	}
	_, err = io.WriteString(stdout, cli.FormatIngestStatus(opts.Repository, &run))
	return err
}

func runIngestPR(opts cli.IngestPRArgs, stdout io.Writer, logger *slog.Logger) error {
	if strings.TrimSpace(opts.Actor) == "" {
		return fmt.Errorf("validation_error: --actor or MEMLORE_ACTOR is required")
	}
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	scope, err := domain.NewScope(domain.ScopeKindRepository, opts.Repository)
	if err != nil {
		return err
	}
	token := githubhttp.TokenFromEnv()
	handler := commands.NewIngestPullRequestsHandler(begin, clock.SystemClock{}, githubhttp.NewReader("", token, nil))
	logger.Info("memlore ingest pr started", "repository_id", opts.Repository)
	run, err := handler.Handle(context.Background(), commands.IngestPullRequestsCommand{
		Scope:   scope,
		ActorID: opts.Actor,
		PR:      opts.PR,
		MaxPRs:  opts.MaxPRs,
	})
	if err != nil {
		logger.Error("memlore ingest pr failed", "repository_id", opts.Repository, "error", err)
		return err
	}
	if run.Status == domain.IngestRunFailed {
		logger.Error("memlore ingest pr failed", "repository_id", opts.Repository, "run_id", run.ID, "error", run.ErrorSummary)
	} else {
		logger.Info("memlore ingest pr completed",
			"repository_id", opts.Repository,
			"run_id", run.ID,
			"prs_seen", run.PRsSeen,
			"candidates_stored", run.CandidatesStored,
		)
	}
	_, err = io.WriteString(stdout, cli.FormatPRIngestStatus(opts.Repository, &run))
	return err
}

func runIngestADR(opts cli.IngestADRArgs, stdout io.Writer, logger *slog.Logger) error {
	if strings.TrimSpace(opts.Actor) == "" {
		return fmt.Errorf("validation_error: --actor or MEMLORE_ACTOR is required")
	}
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	scope, err := domain.NewScope(domain.ScopeKindRepository, opts.Repository)
	if err != nil {
		return err
	}
	handler := commands.NewIngestADRsHandler(begin, clock.SystemClock{}, fsadr.NewReader())
	logger.Info("memlore ingest adr started", "repository_id", opts.Repository, "path", opts.Path)
	run, err := handler.Handle(context.Background(), commands.IngestADRsCommand{
		Scope:     scope,
		Path:      opts.Path,
		ActorID:   opts.Actor,
		ExtraDirs: opts.ADRDirs,
	})
	if err != nil {
		logger.Error("memlore ingest adr failed", "repository_id", opts.Repository, "error", err)
		return err
	}
	if run.Status == domain.IngestRunFailed {
		logger.Error("memlore ingest adr failed", "repository_id", opts.Repository, "run_id", run.ID, "error", run.ErrorSummary)
	} else {
		logger.Info("memlore ingest adr completed",
			"repository_id", opts.Repository,
			"run_id", run.ID,
			"files_seen", run.FilesSeen,
			"lore_stored", run.LoreStored,
		)
	}
	_, err = io.WriteString(stdout, cli.FormatADRIngestStatus(opts.Repository, &run))
	return err
}

func runIngestStatus(opts cli.IngestStatusArgs, stdout io.Writer, _ *slog.Logger) error {
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	scope, err := domain.NewScope(domain.ScopeKindRepository, opts.Repository)
	if err != nil {
		return err
	}
	runs, err := queries.NewListIngestRunsHandler(begin).Handle(context.Background(), queries.ListIngestRunsQuery{Scope: scope})
	if err != nil {
		return err
	}
	if opts.Kind == "pr" {
		prRuns, err := queries.NewListPRIngestRunsHandler(begin).Handle(context.Background(), queries.ListPRIngestRunsQuery{Scope: scope})
		if err != nil {
			return err
		}
		var latest *domain.PRIngestRun
		if len(prRuns) > 0 {
			latest = &prRuns[0]
		}
		_, err = io.WriteString(stdout, cli.FormatPRIngestStatus(opts.Repository, latest))
		return err
	}
	if opts.Kind == "adr" {
		adrRuns, err := queries.NewListADRIngestRunsHandler(begin).Handle(context.Background(), queries.ListADRIngestRunsQuery{Scope: scope})
		if err != nil {
			return err
		}
		var latest *domain.ADRIngestRun
		if len(adrRuns) > 0 {
			latest = &adrRuns[0]
		}
		_, err = io.WriteString(stdout, cli.FormatADRIngestStatus(opts.Repository, latest))
		return err
	}
	var latest *domain.IngestRun
	if len(runs) > 0 {
		latest = &runs[0]
	}
	_, err = io.WriteString(stdout, cli.FormatIngestStatus(opts.Repository, latest))
	return err
}

func runReviewList(opts cli.ReviewListArgs, stdout io.Writer, logger *slog.Logger) error {
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	scope, err := domain.NewScope(domain.ScopeKindRepository, opts.Repository)
	if err != nil {
		return err
	}
	items, err := queries.NewListReviewQueueHandler(begin).Handle(context.Background(), queries.ListReviewQueueQuery{Scope: scope})
	if err != nil {
		return err
	}
	logger.Info("memlore review list", "repository_id", opts.Repository, "pending", len(items))
	_, err = io.WriteString(stdout, cli.FormatReviewList(opts.Repository, items))
	return err
}

func runReviewAccept(opts cli.ReviewAcceptArgs, stdout io.Writer, logger *slog.Logger) error {
	if strings.TrimSpace(opts.Actor) == "" {
		return fmt.Errorf("validation_error: --actor or MEMLORE_ACTOR is required")
	}
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	cmd := commands.AcceptReviewCommand{EntryID: opts.ID, ActorID: opts.Actor}
	if opts.HasEdit {
		stmt := opts.Statement
		cmd.Statement = &stmt
	}
	logger.Info("memlore review accept started", "id", opts.ID, "actor_id", opts.Actor)
	succ, err := commands.NewAcceptReviewHandler(begin, clock.SystemClock{}).Handle(context.Background(), cmd)
	if err != nil {
		logger.Error("memlore review accept failed", "id", opts.ID, "error", err)
		return err
	}
	logger.Info("memlore review accept completed", "id", opts.ID, "successor_id", succ.ID, "origin", succ.Origin)
	_, err = io.WriteString(stdout, cli.FormatReviewAccept(succ))
	return err
}

func runReviewReject(opts cli.ReviewRejectArgs, stdout io.Writer, logger *slog.Logger) error {
	if strings.TrimSpace(opts.Actor) == "" {
		return fmt.Errorf("validation_error: --actor or MEMLORE_ACTOR is required")
	}
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	logger.Info("memlore review reject started", "id", opts.ID, "actor_id", opts.Actor)
	got, err := commands.NewRejectReviewHandler(begin, clock.SystemClock{}).Handle(context.Background(), commands.RejectReviewCommand{
		EntryID: opts.ID,
		ActorID: opts.Actor,
	})
	if err != nil {
		logger.Error("memlore review reject failed", "id", opts.ID, "error", err)
		return err
	}
	logger.Info("memlore review reject completed", "id", opts.ID, "status", got.Status)
	_, err = io.WriteString(stdout, cli.FormatReviewReject(got.ID, got.Status))
	return err
}

func runDecisionCreate(opts cli.DecisionCreateArgs, stdout io.Writer, logger *slog.Logger) error {
	if strings.TrimSpace(opts.Actor) == "" {
		return fmt.Errorf("validation_error: --actor or MEMLORE_ACTOR is required")
	}
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	scope, err := domain.NewScope(domain.ScopeKindRepository, opts.Repository)
	if err != nil {
		return err
	}
	evidence, err := cli.ParseEvidenceFlags(opts.Evidence)
	if err != nil {
		return err
	}
	decided, err := cli.ParseDecisionDate(opts.Date)
	if err != nil {
		return err
	}
	logger.Info("memlore decision create started", "repository_id", opts.Repository, "actor_id", opts.Actor)
	got, err := commands.NewCreateDecisionHandler(begin, clock.SystemClock{}).Handle(context.Background(), commands.CreateDecisionCommand{
		Scope:              scope,
		Question:           opts.Question,
		Choice:             opts.Choice,
		Rationale:          opts.Rationale,
		Alternatives:       cli.DecisionAlternativesFromArgs(opts.Alternatives),
		Consequences:       opts.Consequences,
		Owner:              opts.Owner,
		DecidedAt:          decided,
		AffectedComponents: opts.Components,
		Evidence:           evidence,
		ActorID:            opts.Actor,
	})
	if err != nil {
		logger.Error("memlore decision create failed", "repository_id", opts.Repository, "error", err)
		return err
	}
	logger.Info("memlore decision create completed", "id", got.ID, "source", got.SourceKind)
	_, err = io.WriteString(stdout, cli.FormatDecision(got))
	return err
}

func runDecisionGet(opts cli.DecisionGetArgs, stdout io.Writer, logger *slog.Logger) error {
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	got, err := queries.NewGetDecisionHandler(begin).Handle(context.Background(), queries.GetDecisionQuery{ID: opts.ID})
	if err != nil {
		logger.Error("memlore decision get failed", "id", opts.ID, "error", err)
		return err
	}
	logger.Info("memlore decision get", "id", got.ID, "source", got.SourceKind)
	_, err = io.WriteString(stdout, cli.FormatDecision(got))
	return err
}

func runDecisionList(opts cli.DecisionListArgs, stdout io.Writer, logger *slog.Logger) error {
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	scope, err := domain.NewScope(domain.ScopeKindRepository, opts.Repository)
	if err != nil {
		return err
	}
	items, err := queries.NewListDecisionsHandler(begin).Handle(context.Background(), queries.ListDecisionsQuery{Scope: scope})
	if err != nil {
		logger.Error("memlore decision list failed", "repository_id", opts.Repository, "error", err)
		return err
	}
	logger.Info("memlore decision list", "repository_id", opts.Repository, "count", len(items))
	_, err = io.WriteString(stdout, cli.FormatDecisionList(opts.Repository, items))
	return err
}

func runDecisionSupersede(opts cli.DecisionSupersedeArgs, stdout io.Writer, logger *slog.Logger) error {
	if strings.TrimSpace(opts.Actor) == "" {
		return fmt.Errorf("validation_error: --actor or MEMLORE_ACTOR is required")
	}
	dsn := envOr("MEMLORE_POSTGRES_DSN", "postgresql://memlore:memlore@localhost:15432/memlore")
	dsn = bootstrap.NormalizePostgresDSN(dsn)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	evidence, err := cli.ParseEvidenceFlags(opts.Evidence)
	if err != nil {
		return err
	}
	decided, err := cli.ParseDecisionDate(opts.Date)
	if err != nil {
		return err
	}
	logger.Info("memlore decision supersede started", "id", opts.ID, "actor_id", opts.Actor)
	got, err := commands.NewSupersedeDecisionHandler(begin, clock.SystemClock{}).Handle(context.Background(), commands.SupersedeDecisionCommand{
		PredecessorID:      opts.ID,
		Question:           opts.Question,
		Choice:             opts.Choice,
		Rationale:          opts.Rationale,
		Alternatives:       cli.DecisionAlternativesFromArgs(opts.Alternatives),
		Consequences:       opts.Consequences,
		Owner:              opts.Owner,
		DecidedAt:          decided,
		AffectedComponents: opts.Components,
		Evidence:           evidence,
		ActorID:            opts.Actor,
	})
	if err != nil {
		logger.Error("memlore decision supersede failed", "id", opts.ID, "error", err)
		return err
	}
	logger.Info("memlore decision supersede completed", "id", got.ID, "predecessor_id", opts.ID)
	_, err = io.WriteString(stdout, cli.FormatDecision(got))
	return err
}

func newAuthService() (*appauth.Service, error) {
	cfg := appauth.ConfigFromEnv()
	verifier, err := authadapter.NewVerifier(cfg)
	if err != nil {
		return nil, fmt.Errorf("auth config: %w", err)
	}
	return appauth.NewService(cfg, verifier), nil
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
