package graphclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/memlore/memlore/internal/application/ports"
)

// Client calls graph-service over HTTP and implements ports.KnowledgeGraph.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a graph-service HTTP client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

type healthResponse struct {
	Status string `json:"status"`
	Neo4j  string `json:"neo4j"`
}

type scopeJSON struct {
	Kind string `json:"kind"`
	Key  string `json:"key"`
}

type episodeRequest struct {
	Statement      string         `json:"statement"`
	Scope          scopeJSON      `json:"scope"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	ProvenanceRefs []string       `json:"provenance_refs,omitempty"`
	ReferenceTime  *time.Time     `json:"reference_time,omitempty"`
}

type episodeResponse struct {
	EpisodeID string `json:"episode_id"`
}

type searchRequest struct {
	Query string     `json:"query"`
	Scope *scopeJSON `json:"scope,omitempty"`
	Limit int        `json:"limit,omitempty"`
}

type graphFactJSON struct {
	ID             string     `json:"id"`
	Statement      string     `json:"statement"`
	Score          float64    `json:"score"`
	Scope          *scopeJSON `json:"scope"`
	ProvenanceRefs []string   `json:"provenance_refs"`
}

type searchResponse struct {
	Results []graphFactJSON `json:"results"`
}

type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Health checks graph-service and Neo4j connectivity.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return decodeError(resp)
	}
	var body healthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}
	if body.Neo4j != "ok" {
		return fmt.Errorf("neo4j status: %s", body.Neo4j)
	}
	return nil
}

// IngestEpisode posts an episode to graph-service.
func (c *Client) IngestEpisode(ctx context.Context, req ports.EpisodeIngestRequest) (string, error) {
	payload := episodeRequest{
		Statement:      req.Statement,
		Scope:          scopeJSON{Kind: req.Scope.Kind, Key: req.Scope.Key},
		Metadata:       req.Metadata,
		ProvenanceRefs: req.ProvenanceRefs,
		ReferenceTime:  req.ReferenceTime,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/episodes",
		bytes.NewReader(raw),
	)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", decodeError(resp)
	}
	var body episodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.EpisodeID == "" {
		return "", errors.New("empty episode_id in response")
	}
	return body.EpisodeID, nil
}

// Search queries graph-service for facts.
func (c *Client) Search(ctx context.Context, req ports.SearchRequest) ([]ports.GraphFact, error) {
	payload := searchRequest{Query: req.Query, Limit: req.Limit}
	if req.Scope != nil {
		payload.Scope = &scopeJSON{Kind: req.Scope.Kind, Key: req.Scope.Key}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/search",
		bytes.NewReader(raw),
	)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, decodeError(resp)
	}
	var body searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	out := make([]ports.GraphFact, 0, len(body.Results))
	for _, item := range body.Results {
		out = append(out, toGraphFact(item))
	}
	return out, nil
}

// GetFact retrieves a fact by id.
func (c *Client) GetFact(ctx context.Context, id string) (ports.GraphFact, error) {
	path := c.baseURL + "/facts/" + url.PathEscape(id)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return ports.GraphFact{}, err
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ports.GraphFact{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ports.GraphFact{}, errors.New("fact not found")
	}
	if resp.StatusCode != http.StatusOK {
		return ports.GraphFact{}, decodeError(resp)
	}
	var item graphFactJSON
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return ports.GraphFact{}, err
	}
	return toGraphFact(item), nil
}

func decodeError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var envelope errorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error.Message != "" {
		return fmt.Errorf("%s: %s", envelope.Error.Code, envelope.Error.Message)
	}
	return fmt.Errorf("graph-service status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

func toGraphFact(item graphFactJSON) ports.GraphFact {
	var scope *ports.GraphScope
	if item.Scope != nil {
		scope = &ports.GraphScope{Kind: item.Scope.Kind, Key: item.Scope.Key}
	}
	return ports.GraphFact{
		ID:             item.ID,
		Statement:      item.Statement,
		Score:          item.Score,
		Scope:          scope,
		ProvenanceRefs: item.ProvenanceRefs,
	}
}

// Ensure Client implements KnowledgeGraph at compile time.
var _ ports.KnowledgeGraph = (*Client)(nil)
