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
	graph := &memory.KnowledgeGraph{}
	server := mcpadapter.NewServer(begin, fixed, graph, "test", nil)

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

func TestListToolsIsExactlyNineLoreTools(t *testing.T) {
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
		"memlore.remember":         {},
		"memlore.get":              {},
		"memlore.verify":           {},
		"memlore.explain":          {},
		"memlore.search":           {},
		"memlore.knowledge_search": {},
		"memlore.get_for_task":     {},
		"memlore.invalidate":       {},
		"memlore.supersede":        {},
	}
	if len(names) != 9 {
		t.Fatalf("tool count = %d, want 9: %v", len(names), names)
	}
	for _, name := range names {
		if _, ok := want[name]; !ok {
			t.Fatalf("unexpected tool %q", name)
		}
	}
	joined := strings.ToLower(strings.Join(names, " "))
	for _, forbidden := range []string{"graphiti", "neo4j"} {
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
	if explainPayload["trust_band"] == nil || explainPayload["trust_band"] == "" {
		t.Fatal("explain missing trust_band")
	}
	if _, ok := explainPayload["authority_factors"].(map[string]any); !ok {
		t.Fatalf("explain authority_factors = %v", explainPayload["authority_factors"])
	}
	breakdown, ok := explainPayload["factor_breakdown"].([]any)
	if !ok || len(breakdown) == 0 {
		t.Fatalf("explain factor_breakdown = %v", explainPayload["factor_breakdown"])
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

func TestKnowledgeSearchMCPContract(t *testing.T) {
	session, _ := testSession(t)
	callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Use outbox for payments.",
		"scope":     map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
		"actor_id":  "alice",
	})

	result := callTool(t, session, "memlore.knowledge_search", map[string]any{
		"query":    "payment outbox",
		"scope":    map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
		"actor_id": "alice",
	})
	if result.IsError {
		t.Fatalf("knowledge_search failed: %s", toolText(result))
	}
	payload := structuredContent(t, result)
	if payload["query"] != "payment outbox" {
		t.Fatalf("query = %v", payload["query"])
	}
	governance := payload["governance"].(map[string]any)
	if len(governance["items"].([]any)) != 1 {
		t.Fatalf("governance = %v", governance)
	}
	graph := payload["graph"].(map[string]any)
	if graph["items"] == nil {
		t.Fatal("graph.items should not be null")
	}
	for _, forbidden := range []string{"group_id", "EntityEdge", "graphiti"} {
		raw, _ := json.Marshal(payload)
		if strings.Contains(strings.ToLower(string(raw)), forbidden) {
			t.Fatalf("payload contains forbidden key %q", forbidden)
		}
	}

	missingActor := callTool(t, session, "memlore.knowledge_search", map[string]any{
		"query": "payment outbox",
	})
	if !missingActor.IsError {
		t.Fatal("expected validation error for missing actor_id")
	}
}

func TestGetForTaskMCPContract(t *testing.T) {
	session, _ := testSession(t)
	callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Use outbox for payments.",
		"scope":     map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
		"actor_id":  "alice",
	})

	result := callTool(t, session, "memlore.get_for_task", map[string]any{
		"task":     "Implement payment outbox",
		"query":    "payment outbox",
		"scope":    map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
		"actor_id": "alice",
	})
	if result.IsError {
		t.Fatalf("get_for_task failed: %s", toolText(result))
	}
	payload := structuredContent(t, result)
	if payload["task"] != "Implement payment outbox" {
		t.Fatalf("task = %v", payload["task"])
	}
	items := payload["items"].([]any)
	if len(items) < 1 {
		t.Fatalf("items = %v", items)
	}
	meta := payload["meta"].(map[string]any)
	if meta["items_included"] == nil {
		t.Fatal("missing meta.items_included")
	}
	if _, ok := payload["conflicts"]; !ok {
		t.Fatal("missing conflicts field")
	}
	first := items[0].(map[string]any)
	if first["trust_band"] == nil || first["trust_band"] == "" {
		t.Fatal("get_for_task missing trust_band")
	}
}

func TestTemporalFilterAndConflictsMCPContract(t *testing.T) {
	session, _ := testSession(t)
	scope := map[string]string{"kind": "repository", "key": "r1"}
	a := callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Use blue-green deploys",
		"scope":     scope,
		"actor_id":  "alice",
	})
	b := callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Use rolling deploys",
		"scope":     scope,
		"actor_id":  "alice",
	})
	old := callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Legacy deploy rule",
		"scope":     scope,
		"actor_id":  "alice",
	})
	oldID := structuredContent(t, old)["id"].(string)
	_ = callTool(t, session, "memlore.supersede", map[string]any{
		"id":        oldID,
		"statement": "Successor deploy rule",
		"actor_id":  "alice",
	})

	search := callTool(t, session, "memlore.search", map[string]any{"scope": scope})
	searchItems := structuredContent(t, search)["items"].([]any)
	for _, raw := range searchItems {
		if raw.(map[string]any)["id"].(string) == oldID {
			t.Fatal("default search must omit superseded")
		}
	}

	staleSearch := callTool(t, session, "memlore.search", map[string]any{
		"scope":         scope,
		"include_stale": true,
	})
	foundStale := false
	for _, raw := range structuredContent(t, staleSearch)["items"].([]any) {
		if raw.(map[string]any)["id"].(string) == oldID {
			foundStale = true
		}
	}
	if !foundStale {
		t.Fatal("include_stale search must return superseded")
	}

	got := callTool(t, session, "memlore.get", map[string]any{"id": oldID})
	if got.IsError {
		t.Fatalf("get stale failed: %s", toolText(got))
	}

	packet := callTool(t, session, "memlore.get_for_task", map[string]any{
		"task":     "deploy",
		"scope":    scope,
		"actor_id": "alice",
	})
	if packet.IsError {
		t.Fatalf("get_for_task failed: %s", toolText(packet))
	}
	payload := structuredContent(t, packet)
	for _, raw := range payload["items"].([]any) {
		if raw.(map[string]any)["id"].(string) == oldID {
			t.Fatal("get_for_task must omit superseded from items")
		}
	}
	conflicts := payload["conflicts"].([]any)
	if len(conflicts) < 1 {
		t.Fatalf("expected conflicts, got %v", conflicts)
	}
	_ = a
	_ = b
}

func TestInvalidateAndSupersedeMCPContract(t *testing.T) {
	session, _ := testSession(t)
	created := callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Use the outbox.",
		"scope":     map[string]string{"kind": "repository", "key": "r1"},
		"actor_id":  "alice",
	})
	entryID := structuredContent(t, created)["id"].(string)

	invalidated := callTool(t, session, "memlore.invalidate", map[string]any{
		"id":       entryID,
		"actor_id": "alice",
	})
	if invalidated.IsError {
		t.Fatalf("invalidate failed: %s", toolText(invalidated))
	}
	invPayload := structuredContent(t, invalidated)
	if invPayload["verification_status"] != "invalidated" {
		t.Fatalf("status = %v", invPayload["verification_status"])
	}
	if invPayload["invalidated_by"] != "alice" {
		t.Fatalf("invalidated_by = %v", invPayload["invalidated_by"])
	}

	again := callTool(t, session, "memlore.invalidate", map[string]any{
		"id":       entryID,
		"actor_id": "bob",
	})
	if again.IsError {
		t.Fatalf("re-invalidate failed: %s", toolText(again))
	}
	explained := callTool(t, session, "memlore.explain", map[string]any{"id": entryID})
	audits := structuredContent(t, explained)["audits"].([]any)
	invalidateCount := 0
	for _, item := range audits {
		if item.(map[string]any)["action"] == "invalidate" {
			invalidateCount++
		}
	}
	if invalidateCount != 1 {
		t.Fatalf("invalidate audits = %d (%v)", invalidateCount, audits)
	}

	blank := callTool(t, session, "memlore.invalidate", map[string]any{"id": entryID, "actor_id": ""})
	if !blank.IsError || !strings.Contains(toolText(blank), "validation_error:") {
		t.Fatalf("blank actor invalidate = %v %q", blank.IsError, toolText(blank))
	}

	missing := callTool(t, session, "memlore.invalidate", map[string]any{
		"id":       "00000000-0000-0000-0000-000000000000",
		"actor_id": "alice",
	})
	if !missing.IsError || !strings.Contains(toolText(missing), "not_found:") {
		t.Fatalf("missing invalidate = %v %q", missing.IsError, toolText(missing))
	}

	current := callTool(t, session, "memlore.remember", map[string]any{
		"statement": "Old rule",
		"scope":     map[string]string{"kind": "repository", "key": "r1"},
		"actor_id":  "alice",
	})
	predID := structuredContent(t, current)["id"].(string)
	superseded := callTool(t, session, "memlore.supersede", map[string]any{
		"id":        predID,
		"statement": "New rule",
		"actor_id":  "bob",
		"evidence":  []map[string]string{{"type": "adr", "value": "0003-lifecycle"}},
	})
	if superseded.IsError {
		t.Fatalf("supersede failed: %s", toolText(superseded))
	}
	succPayload := structuredContent(t, superseded)
	if succPayload["statement"] != "New rule" {
		t.Fatalf("successor statement = %v", succPayload["statement"])
	}
	gotPred := structuredContent(t, callTool(t, session, "memlore.get", map[string]any{"id": predID}))
	if gotPred["superseded_by_id"] != succPayload["id"] {
		t.Fatalf("superseded_by_id = %v want %v", gotPred["superseded_by_id"], succPayload["id"])
	}

	predExplain := structuredContent(t, callTool(t, session, "memlore.explain", map[string]any{"id": predID}))
	predAudits := predExplain["audits"].([]any)
	if len(predAudits) < 2 || predAudits[len(predAudits)-1].(map[string]any)["action"] != "supersede" {
		t.Fatalf("predecessor audits = %v", predAudits)
	}

	retry := callTool(t, session, "memlore.supersede", map[string]any{
		"id":        predID,
		"statement": "Another rule",
		"actor_id":  "bob",
	})
	if !retry.IsError || !strings.Contains(toolText(retry), "validation_error:") {
		t.Fatalf("re-supersede = %v %q", retry.IsError, toolText(retry))
	}

	invThenSup := callTool(t, session, "memlore.supersede", map[string]any{
		"id":        entryID,
		"statement": "Should fail",
		"actor_id":  "alice",
	})
	if !invThenSup.IsError || !strings.Contains(toolText(invThenSup), "validation_error:") {
		t.Fatalf("supersede invalidated = %v %q", invThenSup.IsError, toolText(invThenSup))
	}
}
