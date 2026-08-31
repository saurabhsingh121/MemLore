package gitcli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

var _ ports.GitReader = (*Reader)(nil)

// Reader lists commits using the git CLI.
type Reader struct{}

func NewReader() *Reader {
	return &Reader{}
}

func (r *Reader) ListCommits(ctx context.Context, q ports.GitLogQuery) ([]domain.GitCommitSnapshot, error) {
	path := strings.TrimSpace(q.Path)
	if path == "" {
		return nil, &ports.GitNotRepositoryError{}
	}
	if err := r.checkRepo(ctx, path); err != nil {
		return nil, err
	}

	args := []string{"-C", path, "-c", "core.quotepath=false", "log", "HEAD", "--reverse", "--format=%H%x1f%an%x1f%aI%x1f%P%x1f%s%x1f%b%x1e"}
	if q.After != nil && !q.After.IsZero() {
		args = append(args, "--since="+q.After.UTC().Format(time.RFC3339))
	}
	out, err := runGit(ctx, args)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "does not have any commits") ||
			strings.Contains(msg, "bad revision") ||
			strings.Contains(msg, "unknown revision") ||
			strings.Contains(msg, "ambiguous argument 'HEAD'") {
			return []domain.GitCommitSnapshot{}, nil
		}
		return nil, err
	}

	records := strings.Split(string(out), "\x1e")
	commits := make([]domain.GitCommitSnapshot, 0)
	for _, rec := range records {
		rec = strings.TrimSpace(rec)
		if rec == "" {
			continue
		}
		parts := strings.SplitN(rec, "\x1f", 6)
		if len(parts) < 6 {
			continue
		}
		committedAt, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[2]))
		if err != nil {
			committedAt, err = time.Parse(time.RFC3339Nano, strings.TrimSpace(parts[2]))
			if err != nil {
				committedAt = time.Time{}
			}
		}
		parents := strings.Fields(parts[3])
		sha := strings.TrimSpace(parts[0])
		snap := domain.GitCommitSnapshot{
			SHA:         sha,
			Author:      strings.TrimSpace(parts[1]),
			CommittedAt: committedAt.UTC(),
			Subject:     parts[4],
			Body:        strings.TrimRight(parts[5], "\n"),
			ParentCount: len(parents),
		}
		snap.Message = strings.TrimSpace(snap.Subject + "\n\n" + snap.Body)
		paths, err := r.listPaths(ctx, path, sha)
		if err == nil {
			snap.Paths = paths
		}
		commits = append(commits, snap)
		if q.MaxCommits > 0 && len(commits) >= q.MaxCommits {
			break
		}
	}
	return commits, nil
}

func (r *Reader) checkRepo(ctx context.Context, path string) error {
	out, err := runGit(ctx, []string{"-C", path, "rev-parse", "--is-inside-work-tree"})
	if err != nil || strings.TrimSpace(string(out)) != "true" {
		return &ports.GitNotRepositoryError{Path: path}
	}
	return nil
}

func (r *Reader) listPaths(ctx context.Context, repo, sha string) ([]string, error) {
	out, err := runGit(ctx, []string{"-C", repo, "-c", "core.quotepath=false", "diff-tree", "--no-commit-id", "--name-only", "--root", "-r", sha})
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	paths := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	return paths, nil
}

func runGit(ctx context.Context, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return out, nil
}
