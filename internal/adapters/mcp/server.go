package mcpadapter

import (
	"log/slog"

	"github.com/memlore/memlore/internal/application/ports"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewServer registers memlore lore MCP tools on a stdio-ready server.
func NewServer(begin ports.UnitOfWorkFactory, clock ports.Clock, graph ports.KnowledgeGraph, version string, logger *slog.Logger) *sdkmcp.Server {
	opts := &sdkmcp.ServerOptions{}
	if logger != nil {
		opts.Logger = logger
	}
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "memlore", Version: version}, opts)
	tools := NewTools(begin, clock, graph)

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
		Description: "Return lore entry fields plus chronological audits.",
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

	return server
}
