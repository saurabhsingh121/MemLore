package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
)

// ReviewListArgs is parsed input for `memlore review list`.
type ReviewListArgs struct {
	Repository string
}

// ReviewAcceptArgs is parsed input for `memlore review accept`.
type ReviewAcceptArgs struct {
	ID        string
	Statement string
	HasEdit   bool
	Actor     string
}

// ReviewRejectArgs is parsed input for `memlore review reject`.
type ReviewRejectArgs struct {
	ID    string
	Actor string
}

// ParseReviewListArgs parses flags for review list.
func ParseReviewListArgs(args []string) (ReviewListArgs, error) {
	fs := flag.NewFlagSet("review list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repository", "", "repository scope key")
	if err := fs.Parse(args); err != nil {
		return ReviewListArgs{}, err
	}
	key := strings.TrimSpace(*repo)
	if key == "" {
		return ReviewListArgs{}, fmt.Errorf("validation_error: --repository is required")
	}
	return ReviewListArgs{Repository: key}, nil
}

func parseReviewID(args []string, command string) (string, []string, error) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return "", nil, fmt.Errorf("validation_error: %s requires a review item id", command)
	}
	id := strings.TrimSpace(args[0])
	if id == "" {
		return "", nil, fmt.Errorf("validation_error: %s requires a review item id", command)
	}
	return id, args[1:], nil
}

// ParseReviewAcceptArgs parses flags for review accept.
func ParseReviewAcceptArgs(args []string) (ReviewAcceptArgs, error) {
	id, rest, err := parseReviewID(args, "review accept")
	if err != nil {
		return ReviewAcceptArgs{}, err
	}
	fs := flag.NewFlagSet("review accept", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	statement := fs.String("statement", "", "optional edited statement")
	actor := fs.String("actor", "", "acting subject (or MEMLORE_ACTOR)")
	if err := fs.Parse(rest); err != nil {
		return ReviewAcceptArgs{}, err
	}
	out := ReviewAcceptArgs{ID: id, Actor: strings.TrimSpace(*actor), Statement: *statement}
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "statement" {
			out.HasEdit = true
		}
	})
	return out, nil
}

// ParseReviewRejectArgs parses flags for review reject.
func ParseReviewRejectArgs(args []string) (ReviewRejectArgs, error) {
	id, rest, err := parseReviewID(args, "review reject")
	if err != nil {
		return ReviewRejectArgs{}, err
	}
	fs := flag.NewFlagSet("review reject", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	actor := fs.String("actor", "", "acting subject (or MEMLORE_ACTOR)")
	if err := fs.Parse(rest); err != nil {
		return ReviewRejectArgs{}, err
	}
	return ReviewRejectArgs{ID: id, Actor: strings.TrimSpace(*actor)}, nil
}

// FormatReviewList renders pending suggested lore.
func FormatReviewList(repository string, items []queries.SuggestedLoreItem) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Suggested Lore (%s)\n", repository)
	if len(items) == 0 {
		b.WriteString("  (none pending)\n")
		return b.String()
	}
	for _, item := range items {
		fmt.Fprintf(&b, "  id: %s\n", item.Entry.ID)
		fmt.Fprintf(&b, "  statement: %s\n", item.Entry.Statement)
		ev := formatEvidence(item.Entry.Evidence)
		if ev != "" {
			fmt.Fprintf(&b, "  evidence: %s\n", ev)
		}
		fmt.Fprintf(&b, "  source: %s\n", item.SourceType)
		b.WriteString("  [pending]\n")
	}
	return b.String()
}

func formatEvidence(refs []domain.EvidenceReference) string {
	parts := make([]string, 0, len(refs))
	for _, ref := range refs {
		parts = append(parts, fmt.Sprintf("%s %s", ref.Type, ref.Value))
	}
	return strings.Join(parts, ", ")
}

// FormatReviewAccept renders an accepted successor summary.
func FormatReviewAccept(entry domain.LoreEntry) string {
	return fmt.Sprintf("Accepted %s (%s, %s)\n", entry.ID, entry.Origin, entry.VerificationStatus)
}

// FormatReviewReject renders a reject confirmation.
func FormatReviewReject(id string, status domain.ReviewStatus) string {
	return fmt.Sprintf("Rejected %s (%s)\n", id, status)
}
