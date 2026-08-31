package fsadr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

var _ ports.ADRReader = (*Reader)(nil)

// Reader lists markdown ADR files from a local working copy.
type Reader struct{}

func NewReader() *Reader { return &Reader{} }

func (r *Reader) ListADRFiles(_ context.Context, q ports.ADRListQuery) ([]ports.ADRFile, error) {
	root := strings.TrimSpace(q.Path)
	if root == "" {
		return nil, &ports.PathNotDirectoryError{}
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return nil, &ports.PathNotDirectoryError{Path: root}
	}

	dirs := append([]string{}, domain.DefaultADRDirs...)
	for _, d := range q.Dirs {
		if t := strings.TrimSpace(d); t != "" {
			dirs = append(dirs, t)
		}
	}
	seenDir := map[string]bool{}
	seenFile := map[string]bool{}
	out := make([]ports.ADRFile, 0)
	for _, rel := range dirs {
		rel = strings.Trim(strings.ReplaceAll(rel, "\\", "/"), "/")
		if rel == "" || seenDir[rel] {
			continue
		}
		seenDir[rel] = true
		abs := filepath.Join(root, filepath.FromSlash(rel))
		st, err := os.Stat(abs)
		if err != nil || !st.IsDir() {
			continue
		}
		entries, err := os.ReadDir(abs)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if err := collectADR(abs, rel, e, seenFile, &out); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func collectADR(dirAbs, relDir string, e fs.DirEntry, seen map[string]bool, out *[]ports.ADRFile) error {
	name := e.Name()
	if e.IsDir() {
		// one extra level of dated folders, e.g. docs/adr/2026/
		childAbs := filepath.Join(dirAbs, name)
		childRel := relDir + "/" + name
		entries, err := os.ReadDir(childAbs)
		if err != nil {
			return nil
		}
		for _, child := range entries {
			if child.IsDir() {
				continue
			}
			if err := addFile(childAbs, childRel, child, seen, out); err != nil {
				return err
			}
		}
		return nil
	}
	return addFile(dirAbs, relDir, e, seen, out)
}

func addFile(dirAbs, relDir string, e fs.DirEntry, seen map[string]bool, out *[]ports.ADRFile) error {
	name := e.Name()
	if !strings.HasSuffix(strings.ToLower(name), ".md") {
		return nil
	}
	rel := strings.Trim(relDir+"/"+name, "/")
	if seen[rel] {
		return nil
	}
	seen[rel] = true
	body, err := os.ReadFile(filepath.Join(dirAbs, name))
	if err != nil {
		return err
	}
	sum := sha256.Sum256(body)
	*out = append(*out, ports.ADRFile{
		RelativePath: rel,
		Checksum:     hex.EncodeToString(sum[:]),
		Content:      string(body),
	})
	return nil
}
