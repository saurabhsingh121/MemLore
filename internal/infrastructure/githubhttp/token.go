package githubhttp

import (
	"os"
	"strings"
)

// TokenFromEnv returns MEMLORE_GITHUB_TOKEN, else GITHUB_TOKEN. Never log the value.
func TokenFromEnv() string {
	if v := strings.TrimSpace(os.Getenv("MEMLORE_GITHUB_TOKEN")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
}
