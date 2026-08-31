package githubhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

const defaultAPI = "https://api.github.com"

// Reader fetches pull requests from GitHub REST.
type Reader struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewReader builds a GitHub REST reader. token may be empty (calls fail clearly).
func NewReader(baseURL, token string, client *http.Client) *Reader {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultAPI
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &Reader{baseURL: strings.TrimRight(baseURL, "/"), token: strings.TrimSpace(token), httpClient: client}
}

func (r *Reader) ListPullRequests(ctx context.Context, q ports.PullRequestQuery) ([]domain.PullRequestSnapshot, error) {
	if r.token == "" {
		return nil, &ports.GitHubError{Message: "GitHub token is not configured"}
	}
	if q.Number > 0 {
		pr, err := r.getPR(ctx, q.Owner, q.Repo, q.Number)
		if err != nil {
			return nil, err
		}
		return []domain.PullRequestSnapshot{pr}, nil
	}

	var out []domain.PullRequestSnapshot
	page := 1
	for {
		list, err := r.listClosedPage(ctx, q.Owner, q.Repo, page)
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			break
		}
		for _, summary := range list {
			mergedAt, ok := parseMergedAt(summary.MergedAt)
			if !ok {
				continue
			}
			if q.AfterMergedAt != nil && !mergedAt.After(*q.AfterMergedAt) {
				continue
			}
			full, err := r.getPR(ctx, q.Owner, q.Repo, summary.Number)
			if err != nil {
				return nil, err
			}
			out = append(out, full)
			if q.MaxPRs > 0 && len(out) >= q.MaxPRs {
				return out, nil
			}
		}
		if len(list) < 100 {
			break
		}
		page++
		if page > 50 {
			break
		}
	}
	return out, nil
}

type ghUser struct {
	Login string `json:"login"`
	Type  string `json:"type"`
}

type ghPR struct {
	Number   int     `json:"number"`
	NodeID   string  `json:"node_id"`
	Title    string  `json:"title"`
	Body     string  `json:"body"`
	HTMLURL  string  `json:"html_url"`
	MergedAt *string `json:"merged_at"`
	User     ghUser  `json:"user"`
}

type ghFile struct {
	Filename string `json:"filename"`
}

type ghComment struct {
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
	User    ghUser `json:"user"`
}

func (r *Reader) listClosedPage(ctx context.Context, owner, repo string, page int) ([]ghPR, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/pulls?state=closed&sort=updated&direction=asc&per_page=100&page=%d",
		r.baseURL, url.PathEscape(owner), url.PathEscape(repo), page)
	var list []ghPR
	if err := r.getJSON(ctx, u, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *Reader) getPR(ctx context.Context, owner, repo string, number int) (domain.PullRequestSnapshot, error) {
	var raw ghPR
	u := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", r.baseURL, url.PathEscape(owner), url.PathEscape(repo), number)
	if err := r.getJSON(ctx, u, &raw); err != nil {
		return domain.PullRequestSnapshot{}, err
	}
	var files []ghFile
	fu := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/files?per_page=100", r.baseURL, url.PathEscape(owner), url.PathEscape(repo), number)
	if err := r.getJSON(ctx, fu, &files); err != nil {
		return domain.PullRequestSnapshot{}, err
	}
	var review []ghComment
	ru := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/comments?per_page=100", r.baseURL, url.PathEscape(owner), url.PathEscape(repo), number)
	if err := r.getJSON(ctx, ru, &review); err != nil {
		return domain.PullRequestSnapshot{}, err
	}
	var issue []ghComment
	iu := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments?per_page=100", r.baseURL, url.PathEscape(owner), url.PathEscape(repo), number)
	if err := r.getJSON(ctx, iu, &issue); err != nil {
		return domain.PullRequestSnapshot{}, err
	}

	var mergedAt *time.Time
	merged := false
	if t, ok := parseMergedAt(raw.MergedAt); ok {
		mergedAt = &t
		merged = true
	}
	paths := make([]string, 0, len(files))
	for _, f := range files {
		if f.Filename != "" {
			paths = append(paths, f.Filename)
		}
	}
	return domain.PullRequestSnapshot{
		Number:         raw.Number,
		NodeID:         raw.NodeID,
		Owner:          owner,
		Repo:           repo,
		Title:          raw.Title,
		Body:           raw.Body,
		AuthorLogin:    raw.User.Login,
		AuthorIsBot:    strings.EqualFold(raw.User.Type, "Bot"),
		Merged:         merged,
		MergedAt:       mergedAt,
		HTMLURL:        raw.HTMLURL,
		Files:          paths,
		ReviewComments: mapComments(review),
		IssueComments:  mapComments(issue),
	}, nil
}

func mapComments(in []ghComment) []domain.PullRequestComment {
	out := make([]domain.PullRequestComment, 0, len(in))
	for _, c := range in {
		out = append(out, domain.PullRequestComment{
			HTMLURL:     c.HTMLURL,
			Body:        c.Body,
			AuthorLogin: c.User.Login,
			AuthorIsBot: strings.EqualFold(c.User.Type, "Bot"),
		})
	}
	return out
}

func (r *Reader) getJSON(ctx context.Context, rawURL string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "memlore")
	req.Header.Set("Authorization", "Bearer "+r.token)
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return &ports.GitHubError{Message: "GitHub request failed"}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 400 {
		return githubStatusError(resp.StatusCode)
	}
	if dest == nil {
		return nil
	}
	if len(strings.TrimSpace(string(body))) == 0 || string(body) == "null" {
		return nil
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return &ports.GitHubError{Message: "GitHub response was not valid JSON"}
	}
	return nil
}

func githubStatusError(code int) error {
	msg := "GitHub request failed"
	switch code {
	case http.StatusUnauthorized:
		msg = "GitHub credentials were rejected"
	case http.StatusForbidden:
		msg = "GitHub access was denied"
	case http.StatusNotFound:
		msg = "GitHub repository or pull request was not found"
	}
	return &ports.GitHubError{Message: msg, StatusCode: code}
}

func parseMergedAt(raw *string) (time.Time, bool) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(*raw))
	if err != nil {
		return time.Time{}, false
	}
	return t.UTC(), true
}
