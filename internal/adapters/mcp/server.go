package mcpadapter

import (
	"log/slog"

	appauth "github.com/memlore/memlore/internal/application/auth"
	"github.com/memlore/memlore/internal/application/ports"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewServer registers memlore lore MCP tools on a stdio-ready server (local auth mode).
func NewServer(begin ports.UnitOfWorkFactory, clock ports.Clock, graph ports.KnowledgeGraph, version string, logger *slog.Logger) *sdkmcp.Server {
	return NewServerFromTools(NewTools(begin, clock, graph), version, logger)
}

// NewServerFromTools registers the provided tools (with optional Auth configured).
func NewServerFromTools(tools *Tools, version string, logger *slog.Logger) *sdkmcp.Server {
	if tools.Auth == nil {
		tools.Auth = appauth.NewService(appauth.Config{}, nil)
	}
	opts := &sdkmcp.ServerOptions{}
	if logger != nil {
		opts.Logger = logger
	}
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "memlore", Version: version}, opts)

	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "memlore.remember",
		Description: "Store a human-authored scoped lore entry.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: false, IdempotentHint: false},
	}, tools.remember)
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "memlore.get",
		Description: "Fetch a lore entry by id.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, tools.get)
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "memlore.verify",
		Description: "Verify a lore entry (idempotent).",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: false, IdempotentHint: true},
	}, tools.verify)
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "memlore.explain",
		Description: "Return lore entry fields, chronological audits, and explainable authority evaluation.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, tools.explain)
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "memlore.search",
		Description: "List lore entries by exact scope kind and key.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, tools.search)
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "memlore.knowledge_search",
		Description: "Search governance lore and knowledge graph in parallel.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, tools.knowledgeSearch)
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "memlore.get_for_task",
		Description: "Compile token-budgeted context for a coding task.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}, tools.getForTask)
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "memlore.invalidate",
		Description: "Mark a lore entry invalidated without deleting evidence or audits.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: false, IdempotentHint: true},
	}, tools.invalidate)
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "memlore.supersede",
		Description: "Replace a lore entry with a successor while preserving history.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: false, IdempotentHint: false},
	}, tools.supersede)

	return server
}
