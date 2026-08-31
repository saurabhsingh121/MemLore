package fsadr_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/infrastructure/fsadr"
)

func TestReaderListsDefaultAndExtraDirs(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "docs", "adr", "0001.md"), "# One\n")
	mustWrite(t, filepath.Join(root, "architecture", "records", "0002.md"), "# Two\n")
	mustWrite(t, filepath.Join(root, "docs", "adr", "nested", "0003.md"), "# Three\n")

	files, err := fsadr.NewReader().ListADRFiles(context.Background(), ports.ADRListQuery{
		Path: root,
		Dirs: []string{"architecture/records"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, f := range files {
		got[f.RelativePath] = true
		if f.Checksum == "" || f.Content == "" {
			t.Fatalf("incomplete %+v", f)
		}
	}
	if !got["docs/adr/0001.md"] || !got["architecture/records/0002.md"] || !got["docs/adr/nested/0003.md"] {
		t.Fatalf("got %+v", got)
	}
}

func TestReaderMissingPath(t *testing.T) {
	_, err := fsadr.NewReader().ListADRFiles(context.Background(), ports.ADRListQuery{Path: filepath.Join(t.TempDir(), "nope")})
	var pe *ports.PathNotDirectoryError
	if err == nil {
		t.Fatal("expected path error")
	}
	if !errorAs(err, &pe) {
		t.Fatalf("err = %v", err)
	}
}

func TestReaderEmptyDirsSucceed(t *testing.T) {
	root := t.TempDir()
	files, err := fsadr.NewReader().ListADRFiles(context.Background(), ports.ADRListQuery{Path: root})
	if err != nil || len(files) != 0 {
		t.Fatalf("files=%d err=%v", len(files), err)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func errorAs(err error, target **ports.PathNotDirectoryError) bool {
	if pe, ok := err.(*ports.PathNotDirectoryError); ok {
		*target = pe
		return true
	}
	return false
}
