package ports

import (
	"context"
	"fmt"
)

// ADRListQuery selects ADR files from a local working copy.
type ADRListQuery struct {
	Path  string
	Dirs  []string // relative dirs to scan; empty means DefaultADRDirs only
}

// ADRFile is a discovered markdown file plus checksummed body.
type ADRFile struct {
	RelativePath string
	Checksum     string
	Content      string
}

// ADRReader lists ADR files from a local directory tree.
type ADRReader interface {
	ListADRFiles(ctx context.Context, q ADRListQuery) ([]ADRFile, error)
}

// PathNotDirectoryError is returned when path is missing or not a directory.
type PathNotDirectoryError struct {
	Path string
}

func (e *PathNotDirectoryError) Error() string {
	if e.Path == "" {
		return "path is not a directory"
	}
	return fmt.Sprintf("path is not a directory: %s", e.Path)
}
