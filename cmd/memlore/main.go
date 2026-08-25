package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Version is the MemLore Core build version (skeleton).
const Version = "0.1.0-dev"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	os.Exit(run(os.Args, os.Stdout, logger))
}

func run(args []string, stdout io.Writer, logger *slog.Logger) int {
	if len(args) < 2 {
		fmt.Fprintf(stdout, "memlore %s — MemLore Core (Go skeleton)\n", Version)
		fmt.Fprintln(stdout, "usage: memlore version")
		return 0
	}
	switch args[1] {
	case "version":
		fmt.Fprintln(stdout, Version)
		logger.Info("memlore version", "version", Version)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[1])
		return 1
	}
}
