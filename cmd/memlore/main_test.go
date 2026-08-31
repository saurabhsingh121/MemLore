package main

import (
	"bytes"
	"io"
	"log/slog"
	"testing"
)

func TestVersionCommandPrintsVersionAndExitsZero(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"memlore", "version"}, io.Writer(&buf), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if code != 0 {
		t.Fatalf("run(version) = %d, want 0", code)
	}
	out := buf.String()
	if out == "" {
		t.Fatal("expected version output, got empty")
	}
	if !bytes.Contains(buf.Bytes(), []byte(Version)) {
		t.Fatalf("output %q does not contain version %q", out, Version)
	}
}

func TestUnknownSubcommandExitsNonZero(t *testing.T) {
	code := run([]string{"memlore", "unknown"}, io.Discard, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if code == 0 {
		t.Fatal("run(unknown) = 0, want non-zero")
	}
}

func TestProfileWithoutRepositoryExitsNonZero(t *testing.T) {
	code := run([]string{"memlore", "profile"}, io.Discard, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if code == 0 {
		t.Fatal("run(profile) = 0, want non-zero")
	}
}

func TestUsageMentionsProfile(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)
	if !bytes.Contains(buf.Bytes(), []byte("profile --repository")) {
		t.Fatalf("usage missing profile: %s", buf.String())
	}
}

func TestContextWithoutTaskExitsNonZero(t *testing.T) {
	code := run([]string{"memlore", "context", "--repository", "r1"}, io.Discard, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if code == 0 {
		t.Fatal("run(context) = 0, want non-zero")
	}
}

func TestUsageMentionsContext(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)
	if !bytes.Contains(buf.Bytes(), []byte("context --task")) {
		t.Fatalf("usage missing context: %s", buf.String())
	}
}

func TestUsageMentionsIngest(t *testing.T) {
	var buf bytes.Buffer
	printUsage(&buf)
	if !bytes.Contains(buf.Bytes(), []byte("ingest git --repository")) {
		t.Fatalf("usage missing ingest: %s", buf.String())
	}
	if !bytes.Contains(buf.Bytes(), []byte("ingest pr --repository")) {
		t.Fatalf("usage missing ingest pr: %s", buf.String())
	}
	if !bytes.Contains(buf.Bytes(), []byte("ingest adr --repository")) {
		t.Fatalf("usage missing ingest adr: %s", buf.String())
	}
}

func TestIngestPRWithoutRepositoryExitsNonZero(t *testing.T) {
	code := run([]string{"memlore", "ingest", "pr"}, io.Discard, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if code == 0 {
		t.Fatal("run(ingest pr) = 0, want non-zero")
	}
}

func TestIngestGitWithoutRepositoryExitsNonZero(t *testing.T) {
	code := run([]string{"memlore", "ingest", "git"}, io.Discard, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if code == 0 {
		t.Fatal("run(ingest git) = 0, want non-zero")
	}
}

func TestIngestADRWithoutRepositoryExitsNonZero(t *testing.T) {
	code := run([]string{"memlore", "ingest", "adr"}, io.Discard, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if code == 0 {
		t.Fatal("run(ingest adr) = 0, want non-zero")
	}
}
