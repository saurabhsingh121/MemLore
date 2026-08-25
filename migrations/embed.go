package migrations

import "embed"

// Goose contains embedded goose migration SQL files.
//
//go:embed *.sql
var Goose embed.FS
