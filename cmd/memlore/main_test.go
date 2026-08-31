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
