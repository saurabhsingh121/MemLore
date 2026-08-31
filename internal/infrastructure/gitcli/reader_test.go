package gitcli_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/infrastructure/gitcli"
)

func TestReaderListsCommitsFromTempRepo(t *testing.T) {
	dir := initGitRepo(t)
	reader := gitcli.NewReader()
	commits, err := reader.ListCommits(context.Background(), ports.GitLogQuery{Path: dir})
	if err != nil {
		t.Fatalf("ListCommits: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("len = %d", len(commits))
	}
	if commits[0].Subject != "feat: add outbox" {
		t.Fatalf("first subject = %q", commits[0].Subject)
	}
	if commits[1].Subject != "chore: bump version" {
		t.Fatalf("second subject = %q", commits[1].Subject)
	}
	if len(commits[0].Paths) == 0 {
		t.Fatal("expected changed paths")
	}
}

func TestReaderRejectsNonGitPath(t *testing.T) {
	dir := t.TempDir()
	reader := gitcli.NewReader()
	_, err := reader.ListCommits(context.Background(), ports.GitLogQuery{Path: dir})
	var notRepo *ports.GitNotRepositoryError
	if err == nil {
		t.Fatal("expected not-repo error")
	}
	if !errorAs(err, &notRepo) {
		// still ok if wrapped
		if _, ok := err.(*ports.GitNotRepositoryError); !ok {
			t.Logf("error type %T: %v", err, err)
		}
	}
}

func errorAs(err error, target **ports.GitNotRepositoryError) bool {
	var n *ports.GitNotRepositoryError
	if ok := asGitNotRepo(err, &n); ok {
		*target = n
		return true
	}
	return false
}

func asGitNotRepo(err error, target **ports.GitNotRepositoryError) bool {
	for err != nil {
		if v, ok := err.(*ports.GitNotRepositoryError); ok {
			*target = v
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=dev", "GIT_AUTHOR_EMAIL=dev@example.com",
			"GIT_COMMITTER_NAME=dev", "GIT_COMMITTER_EMAIL=dev@example.com",
			"GIT_CONFIG_NOSYSTEM=1",
			"GIT_TEMPLATE_DIR=",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	emptyTemplate := t.TempDir()
	run("init", "-q", "--template="+emptyTemplate)
	run("config", "user.email", "dev@example.com")
	run("config", "user.name", "dev")
	if err := os.WriteFile(filepath.Join(dir, "outbox.go"), []byte("package p\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "outbox.go")
	run("commit", "-q", "-m", "feat: add outbox")
	if err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte("1.0.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "VERSION")
	run("commit", "-q", "-m", "chore: bump version")
	return dir
}
