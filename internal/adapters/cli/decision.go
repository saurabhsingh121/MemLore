package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

// DecisionCreateArgs is parsed input for `memlore decision create`.
type DecisionCreateArgs struct {
	Repository   string
	Question     string
	Choice       string
	Owner        string
	Rationale    string
	Consequences string
	Alternatives []string
	Components   []string
	Evidence     []string
	Date         string
	Actor        string
}

// DecisionGetArgs is parsed input for `memlore decision get`.
type DecisionGetArgs struct {
	ID string
}

// DecisionListArgs is parsed input for `memlore decision list`.
type DecisionListArgs struct {
	Repository string
}

// DecisionSupersedeArgs is parsed input for `memlore decision supersede`.
type DecisionSupersedeArgs struct {
	ID           string
	Question     string
	Choice       string
	Owner        string
	Rationale    string
	Consequences string
	Alternatives []string
	Components   []string
	Evidence     []string
	Date         string
	Actor        string
}

type repeatableStrings []string

func (s *repeatableStrings) String() string { return strings.Join(*s, ",") }
func (s *repeatableStrings) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func parseDecisionFields(fs *flag.FlagSet, args []string) (question, choice, owner, rationale, consequences, date, actor string, alts, comps, evidence []string, err error) {
	q := fs.String("question", "", "problem / question")
	c := fs.String("choice", "", "the decision")
	o := fs.String("owner", "", "decision owner")
	r := fs.String("rationale", "", "optional rationale")
	cons := fs.String("consequences", "", "optional consequences")
	dt := fs.String("date", "", "optional RFC3339 decision date")
	act := fs.String("actor", "", "acting subject (or MEMLORE_ACTOR)")
	var alternatives, components, evid repeatableStrings
	fs.Var(&alternatives, "alternative", "considered alternative (repeatable)")
	fs.Var(&components, "component", "affected component (repeatable)")
	fs.Var(&evid, "evidence", "evidence as type:value (repeatable)")
	if err := fs.Parse(args); err != nil {
		return "", "", "", "", "", "", "", nil, nil, nil, err
	}
	return strings.TrimSpace(*q), strings.TrimSpace(*c), strings.TrimSpace(*o),
		strings.TrimSpace(*r), strings.TrimSpace(*cons), strings.TrimSpace(*dt), strings.TrimSpace(*act),
		[]string(alternatives), []string(components), []string(evid), nil
}

// ParseDecisionCreateArgs parses flags for decision create.
func ParseDecisionCreateArgs(args []string) (DecisionCreateArgs, error) {
	fs := flag.NewFlagSet("decision create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repository", "", "repository scope key")
	question, choice, owner, rationale, consequences, date, actor, alts, comps, evid, err := parseDecisionFields(fs, args)
	if err != nil {
		return DecisionCreateArgs{}, err
	}
	key := strings.TrimSpace(*repo)
	if key == "" {
		return DecisionCreateArgs{}, fmt.Errorf("validation_error: --repository is required")
	}
	if question == "" || choice == "" || owner == "" {
		return DecisionCreateArgs{}, fmt.Errorf("validation_error: --question, --choice, and --owner are required")
	}
	return DecisionCreateArgs{
		Repository: key, Question: question, Choice: choice, Owner: owner,
		Rationale: rationale, Consequences: consequences, Alternatives: alts,
		Components: comps, Evidence: evid, Date: date, Actor: actor,
	}, nil
}

// ParseDecisionGetArgs parses `decision get <id>`.
func ParseDecisionGetArgs(args []string) (DecisionGetArgs, error) {
	id, _, err := parseReviewID(args, "decision get")
	if err != nil {
		return DecisionGetArgs{}, err
	}
	return DecisionGetArgs{ID: id}, nil
}

// ParseDecisionListArgs parses flags for decision list.
func ParseDecisionListArgs(args []string) (DecisionListArgs, error) {
	fs := flag.NewFlagSet("decision list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repository", "", "repository scope key")
	if err := fs.Parse(args); err != nil {
		return DecisionListArgs{}, err
	}
	key := strings.TrimSpace(*repo)
	if key == "" {
		return DecisionListArgs{}, fmt.Errorf("validation_error: --repository is required")
	}
	return DecisionListArgs{Repository: key}, nil
}

// ParseDecisionSupersedeArgs parses `decision supersede <id>`.
func ParseDecisionSupersedeArgs(args []string) (DecisionSupersedeArgs, error) {
	id, rest, err := parseReviewID(args, "decision supersede")
	if err != nil {
		return DecisionSupersedeArgs{}, err
	}
	fs := flag.NewFlagSet("decision supersede", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	question, choice, owner, rationale, consequences, date, actor, alts, comps, evid, err := parseDecisionFields(fs, rest)
	if err != nil {
		return DecisionSupersedeArgs{}, err
	}
	if question == "" || choice == "" || owner == "" {
		return DecisionSupersedeArgs{}, fmt.Errorf("validation_error: --question, --choice, and --owner are required")
	}
	return DecisionSupersedeArgs{
		ID: id, Question: question, Choice: choice, Owner: owner,
		Rationale: rationale, Consequences: consequences, Alternatives: alts,
		Components: comps, Evidence: evid, Date: date, Actor: actor,
	}, nil
}

// ParseEvidenceFlags converts type:value strings into evidence refs.
func ParseEvidenceFlags(values []string) ([]domain.EvidenceReference, error) {
	out := make([]domain.EvidenceReference, 0, len(values))
	for _, raw := range values {
		typeName, value, ok := strings.Cut(strings.TrimSpace(raw), ":")
		if !ok || strings.TrimSpace(typeName) == "" || strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("validation_error: --evidence must be type:value")
		}
		et, err := domain.ParseEvidenceType(strings.TrimSpace(typeName))
		if err != nil {
			return nil, err
		}
		ref, err := domain.NewEvidenceReference(et, strings.TrimSpace(value))
		if err != nil {
			return nil, err
		}
		out = append(out, ref)
	}
	return out, nil
}

// ParseDecisionDate parses optional RFC3339 date.
func ParseDecisionDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	got, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("validation_error: --date must be RFC3339")
	}
	return got.UTC(), nil
}

func alternativesFromLabels(labels []string) []domain.DecisionAlternative {
	out := make([]domain.DecisionAlternative, 0, len(labels))
	for _, label := range labels {
		out = append(out, domain.DecisionAlternative{Label: label})
	}
	return out
}

// FormatDecisionList renders current decisions.
func FormatDecisionList(repository string, items []domain.Decision) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Decisions (%s)\n", repository)
	if len(items) == 0 {
		b.WriteString("  (none current)\n")
		return b.String()
	}
	for _, item := range items {
		b.WriteString(formatDecision(item))
	}
	return b.String()
}

// FormatDecision renders one Decision.
func FormatDecision(d domain.Decision) string {
	return formatDecision(d)
}

func formatDecision(d domain.Decision) string {
	var b strings.Builder
	fmt.Fprintf(&b, "  id: %s\n", d.ID)
	if strings.TrimSpace(d.Question) != "" {
		fmt.Fprintf(&b, "  question: %s\n", d.Question)
	}
	fmt.Fprintf(&b, "  decision: %s\n", d.Choice)
	fmt.Fprintf(&b, "  owner: %s\n", d.Owner)
	fmt.Fprintf(&b, "  source: %s\n", d.SourceKind)
	ev := formatEvidence(d.Evidence)
	if ev != "" {
		fmt.Fprintf(&b, "  evidence: %s\n", ev)
	}
	if d.Current {
		b.WriteString("  [current]\n")
	} else {
		b.WriteString("  [superseded]\n")
	}
	return b.String()
}

// DecisionAlternativesFromArgs maps CLI alternative labels.
func DecisionAlternativesFromArgs(labels []string) []domain.DecisionAlternative {
	return alternativesFromLabels(labels)
}
