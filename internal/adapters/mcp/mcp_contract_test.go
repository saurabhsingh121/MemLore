package mcpadapter_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	mcpadapter "github.com/memlore/memlore/internal/adapters/mcp"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Characterization: tests/contract/test_mcp_*.py

func testSession(t *testing.T) (*sdkmcp.ClientSession, *memory.UnitOfWork) {
	t.Helper()
	ctx := context.Background()
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	fixed := clock.FixedClock{Instant: time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)}
	server := mcpadapter.NewServer(begin, fixed, "test", nil)

	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client", Version: "test"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session, uow
}

func toolText(result *sdkmcp.CallToolResult) string {
	var parts []string
	for _, block := range result.Content {
		if text, ok := block.(*sdkmcp.TextContent); ok {
			parts = append(parts, text.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func callTool(t *testing.T, session *sdkmcp.ClientSession, name string, args map[string]any) *sdkmcp.CallToolResult {
	t.Helper()
	result, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool(%s): %v", name, err)
	}
	return result
}

func structuredContent(t *testing.T, result *sdkmcp.CallToolResult) map[string]any {
	t.Helper()
	if result.StructuredContent == nil {
		t.Fatal("expected structured content")
	}
	payload, ok := result.StructuredContent.(map[string]any)
	if !ok {
		raw, err := json.Marshal(result.StructuredContent)
		if err != nil {
			t.Fatalf("marshal structured content: %v", err)
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatalf("unmarshal structured content: %v", err)
		}
	}
	return payload
}

func TestListToolsIsExactlyFiveLoreTools(t *testing.T) {
	session, _ := testSession(t)
	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	names := make([]string, 0, len(listed.Tools))
	for _, tool := range listed.Tools {
		names = append(names, tool.Name)
	}
	want := map[string]struct{}{
		"memlore.remember": {},
		"memlore.get":      {},
		"memlore.verify":   {},
		"memlore.explain":  {},
		"memlore.search":   {},
	}
	if len(names) != 5 {
		t.Fatalf("tool count = %d, want 5: %v", len(names), names)
	}
	for _, name := range names {
		if _, ok := want[name]; !ok {
			t.Fatalf("unexpected tool %q", name)
		}
	}
	joined := strings.ToLower(strings.Join(names, " "))
	for _, forbidden := range []string{"graphiti", "neo4j", "get_for_task", "supersede", "invalidate"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("forbidden tool name fragment %q in %v", forbidden, names)
		}
	}
}

func TestRememberSuccessAndDuplicate(t *testing.T) {
	session, _ := testSession(t)
	args := map[string]any{
		"statement": "Payment events must use the transactional outbox.",
		"scope":     map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
		"actor_id":  "alice",
		"evidence":  []map[string]string{{"type": "adr", "value": "0001-dual-plane"}},
	}
	first := callTool(t, session, "memlore.remember", args)
	if first.IsError {
		t.Fatalf("remember failed: %s", toolText(first))
	}
	payload := structuredContent(t, first)
	if payload["origin"] != "human_authored" {
		t.Fatalf("origin = %v", payload["origin"])
	}
	if payload["verification_status"] != "unverified" {
		t.Fatalf("status = %v", payload["verification_status"])
	}

	second := callTool(t, session, "memlore.remember", args)
	if second.IsError {
		t.Fatalf("duplicate remember failed: %s", toolText(second))
	}
	secondPayload := structuredContent(t, second)
	if secondPayload["id"] == payload["id"] {
		t.Fatal("duplicate remember should create a new id")
	}
}

func TestRememberMissingOrBlankActorIsValidationError(t *testing.T) {
	session, uow := testSession(t)
	missing := callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Rule",
		"scope":     map[string]string{"kind": "team", "key": "t1"},
	})
	if !missing.IsError {
		t.Fatal("expected validation error for missing actor")
	}
	if uow.LoreEntries().(*memory.LoreRepository).Len() != 0 {
		t.Fatal("missing actor should not persist")
	}

	blank := callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Rule",
		"scope":     map[string]string{"kind": "team", "key": "t1"},
		"actor_id":  "  ",
	})
	if !blank.IsError {
		t.Fatal("expected validation error for blank actor")
	}
	if !strings.Contains(toolText(blank), "validation_error:") {
		t.Fatalf("blank actor text = %q", toolText(blank))
	}
}

func TestGetAndExplainExistingAndUnknown(t *testing.T) {
	session, _ := testSession(t)
	created := callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Prefer explicit actor_id",
		"scope":     map[string]string{"kind": "repository", "key": "r1"},
		"actor_id":  "alice",
	})
	entryID := structuredContent(t, created)["id"].(string)

	got := callTool(t, session, "memlore.get", map[string]any{"id": entryID})
	if got.IsError {
		t.Fatalf("get failed: %s", toolText(got))
	}
	gotPayload := structuredContent(t, got)
	if gotPayload["id"] != entryID {
		t.Fatalf("id = %v", gotPayload["id"])
	}

	explained := callTool(t, session, "memlore.explain", map[string]any{"id": entryID})
	if explained.IsError {
		t.Fatalf("explain failed: %s", toolText(explained))
	}
	explainPayload := structuredContent(t, explained)
	if _, ok := explainPayload["summary"]; ok {
		t.Fatal("explain should not include summary")
	}
	audits := explainPayload["audits"].([]any)
	if len(audits) != 1 || audits[0].(map[string]any)["action"] != "create" {
		t.Fatalf("audits = %v", audits)
	}

	missing := "00000000-0000-0000-0000-000000000000"
	getMissing := callTool(t, session, "memlore.get", map[string]any{"id": missing})
	if !getMissing.IsError || !strings.Contains(toolText(getMissing), "not_found:") {
		t.Fatalf("get missing = %v %q", getMissing.IsError, toolText(getMissing))
	}
	explainMissing := callTool(t, session, "memlore.explain", map[string]any{"id": missing})
	if !explainMissing.IsError || !strings.Contains(toolText(explainMissing), "not_found:") {
		t.Fatalf("explain missing = %v %q", explainMissing.IsError, toolText(explainMissing))
	}
	if strings.Contains(toolText(explainMissing), "Traceback") {
		t.Fatal("explain missing leaked traceback")
	}
}

func TestVerifySuccessIdempotentAndErrors(t *testing.T) {
	session, _ := testSession(t)
	created := callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Rule",
		"scope":     map[string]string{"kind": "team", "key": "core"},
		"actor_id":  "alice",
	})
	entryID := structuredContent(t, created)["id"].(string)

	verified := callTool(t, session, "memlore.verify", map[string]any{"id": entryID, "actor_id": "alice"})
	if verified.IsError {
		t.Fatalf("verify failed: %s", toolText(verified))
	}
	verifiedPayload := structuredContent(t, verified)
	if verifiedPayload["verification_status"] != "verified" {
		t.Fatalf("status = %v", verifiedPayload["verification_status"])
	}
	if verifiedPayload["verified_by"] != "alice" {
		t.Fatalf("verified_by = %v", verifiedPayload["verified_by"])
	}

	again := callTool(t, session, "memlore.verify", map[string]any{"id": entryID, "actor_id": "bob"})
	if again.IsError {
		t.Fatalf("re-verify failed: %s", toolText(again))
	}
	againPayload := structuredContent(t, again)
	if againPayload["verified_by"] != "alice" {
		t.Fatalf("verified_by after re-verify = %v", againPayload["verified_by"])
	}

	blank := callTool(t, session, "memlore.verify", map[string]any{"id": entryID, "actor_id": ""})
	if !blank.IsError || !strings.Contains(toolText(blank), "validation_error:") {
		t.Fatalf("blank actor verify = %v %q", blank.IsError, toolText(blank))
	}
}

func TestSearchExactScopeEmptyAndIncomplete(t *testing.T) {
	session, _ := testSession(t)
	callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Alpha rule",
		"scope":     map[string]string{"kind": "repository", "key": "alpha"},
		"actor_id":  "alice",
	})
	callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Beta rule",
		"scope":     map[string]string{"kind": "repository", "key": "beta"},
		"actor_id":  "alice",
	})

	found := callTool(t, session, "memlore.search", map[string]any{
		"scope": map[string]string{"kind": "repository", "key": "alpha"},
	})
	foundPayload := structuredContent(t, found)
	items := foundPayload["items"].([]any)
	if len(items) != 1 || items[0].(map[string]any)["statement"] != "Alpha rule" {
		t.Fatalf("search alpha = %v", items)
	}

	empty := callTool(t, session, "memlore.search", map[string]any{
		"scope": map[string]string{"kind": "team", "key": "nobody"},
	})
	emptyPayload := structuredContent(t, empty)
	if len(emptyPayload["items"].([]any)) != 0 {
		t.Fatalf("empty search = %v", emptyPayload["items"])
	}

	incomplete := callTool(t, session, "memlore.search", map[string]any{
		"scope": map[string]string{"kind": "team"},
	})
	if !incomplete.IsError {
		t.Fatal("expected validation error for incomplete scope")
	}
	text := toolText(incomplete)
	if !strings.Contains(text, "validation_error:") && !strings.Contains(strings.ToLower(text), "required") {
		t.Fatalf("incomplete scope text = %q", text)
	}
}
