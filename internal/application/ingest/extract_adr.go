package ingest

import (
	"path"
	"regexp"
	"strings"

	"github.com/memlore/memlore/internal/domain"
)

var (
	headingLine       = regexp.MustCompile(`^(#{1,3})\s+(.+?)\s*$`)
	frontMatterStatus = regexp.MustCompile(`(?i)(?m)^status:\s*(.+)$`)
	templateName      = regexp.MustCompile(`(?i)^n+-title\.md$`)
	adrIDToken        = regexp.MustCompile(`(?i)(?:ADR[- ]?)?(\d{1,4}(?:-[a-z0-9]+)?)`)
	mdExt             = regexp.MustCompile(`(?i)\.md$`)
)

// ADRExtractResult is a conservative parse of one ADR file.
type ADRExtractResult struct {
	OK         bool
	SkipReason string
	Snapshot   domain.ADRSnapshot
	Statement  string
}

// ExtractADR parses an ADR file conservatively. It does not invent a decision.
func ExtractADR(relativePath, content string) ADRExtractResult {
	rel := strings.ReplaceAll(strings.TrimSpace(relativePath), "\\", "/")
	base := path.Base(rel)
	if skip, reason := skipADRFilename(base); skip {
		return ADRExtractResult{SkipReason: reason}
	}

	body := strings.ReplaceAll(content, "\r\n", "\n")
	front, rest := splitFrontMatter(body)
	statusRaw := ""
	if m := frontMatterStatus.FindStringSubmatch(front); len(m) == 2 {
		statusRaw = strings.TrimSpace(m[1])
	}

	title, sections := parseADRSections(rest)
	if title == "" {
		title = strings.TrimSuffix(base, path.Ext(base))
		title = strings.ReplaceAll(title, "-", " ")
	}
	if statusRaw == "" {
		statusRaw = firstStatusLine(sections["status"])
	}
	decision := strings.TrimSpace(sections["decision"])
	if decision == "" {
		return ADRExtractResult{SkipReason: domain.IngestSkipNoDecision, Snapshot: domain.ADRSnapshot{
			RelativePath: rel, Title: title, StatusRaw: statusRaw, Body: rest,
		}}
	}
	class := domain.ClassifyADRStatus(statusRaw)
	if class == domain.ADRStatusUnknown && strings.TrimSpace(statusRaw) == "" {
		return ADRExtractResult{SkipReason: domain.IngestSkipUnknownStatus, Snapshot: domain.ADRSnapshot{
			RelativePath: rel, Title: title, Decision: decision, Body: rest,
		}}
	}
	if class == domain.ADRStatusUnknown {
		return ADRExtractResult{SkipReason: domain.IngestSkipUnknownStatus, Snapshot: domain.ADRSnapshot{
			RelativePath: rel, Title: title, StatusRaw: statusRaw, Decision: decision, Body: rest,
		}}
	}
	if class == domain.ADRStatusSkip {
		reason := domain.IngestSkipDraft
		tok := strings.ToLower(strings.Fields(strings.TrimSpace(statusRaw))[0])
		if tok == "rejected" || tok == "withdrawn" {
			reason = domain.IngestSkipRejected
		}
		return ADRExtractResult{SkipReason: reason, Snapshot: domain.ADRSnapshot{
			RelativePath: rel, Title: title, StatusRaw: statusRaw, StatusClass: class, Decision: decision,
		}}
	}

	snap := domain.ADRSnapshot{
		RelativePath: rel,
		Title:        title,
		StatusRaw:    statusRaw,
		StatusClass:  class,
		Decision:     decision,
		Context:      strings.TrimSpace(sections["context"]),
		Alternatives: strings.TrimSpace(sections["alternatives"]),
		Consequences: strings.TrimSpace(sections["consequences"]),
		Supersedes:   parseSupersedes(sections["supersedes"], sections["status"]),
		Components:   parseComponentList(sections["components"]),
		ADRID:        domain.ADRIdentityFromPath(rel),
		Body:         rest,
	}
	statement := buildADRStatement(snap)
	if statement == "" {
		return ADRExtractResult{SkipReason: domain.IngestSkipNoDecision, Snapshot: snap}
	}
	if len(statement) > domain.MaxStatementLength {
		statement = buildADRStatement(domain.ADRSnapshot{Title: snap.Title, Decision: snap.Decision})
		if statement == "" || len(statement) > domain.MaxStatementLength {
			return ADRExtractResult{SkipReason: domain.IngestSkipTooLong, Snapshot: snap}
		}
	}
	snap.StatusClass = class
	return ADRExtractResult{OK: true, Snapshot: snap, Statement: statement}
}

func skipADRFilename(base string) (bool, string) {
	lower := strings.ToLower(base)
	if !mdExt.MatchString(lower) {
		return true, domain.IngestSkipNotADR
	}
	if lower == "readme.md" {
		return true, domain.IngestSkipReadme
	}
	if lower == "template.md" || templateName.MatchString(lower) {
		return true, domain.IngestSkipTemplate
	}
	return false, ""
}

func splitFrontMatter(body string) (front, rest string) {
	body = strings.TrimPrefix(body, "\ufeff")
	if !strings.HasPrefix(body, "---") {
		return "", body
	}
	trimmed := strings.TrimPrefix(body, "---")
	trimmed = strings.TrimPrefix(trimmed, "\n")
	idx := strings.Index(trimmed, "\n---")
	if idx < 0 {
		return "", body
	}
	return trimmed[:idx], strings.TrimLeft(trimmed[idx+len("\n---"):], "\n")
}

func parseADRSections(body string) (title string, sections map[string]string) {
	sections = map[string]string{}
	lines := strings.Split(body, "\n")
	current := ""
	var buf strings.Builder
	flush := func() {
		if current == "" {
			return
		}
		sections[current] += buf.String()
		buf.Reset()
	}
	for _, line := range lines {
		if m := headingLine.FindStringSubmatch(line); len(m) == 3 {
			name := normalizeHeading(m[2])
			if title == "" && m[1] == "#" && !isSectionHeading(name) {
				title = strings.TrimSpace(m[2])
				continue
			}
			if isSectionHeading(name) {
				flush()
				current = canonicalSection(name)
				continue
			}
		}
		if current != "" {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}
	flush()
	return title, sections
}

func normalizeHeading(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	h = strings.Trim(h, ":")
	return h
}

func isSectionHeading(name string) bool {
	return canonicalSection(name) != ""
}

func canonicalSection(name string) string {
	switch name {
	case "status":
		return "status"
	case "context", "context and problem statement":
		return "context"
	case "decision", "decision outcome":
		return "decision"
	case "consequences", "consequence":
		return "consequences"
	case "alternatives", "considered alternatives", "options", "considered options":
		return "alternatives"
	case "supersedes", "supersede", "supercedes":
		return "supersedes"
	case "components", "affected", "affected components":
		return "components"
	default:
		return ""
	}
}

func firstStatusLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

func parseSupersedes(section, status string) []string {
	seen := map[string]bool{}
	out := make([]string, 0)
	add := func(text string) {
		for _, m := range adrIDToken.FindAllStringSubmatch(text, -1) {
			if len(m) < 2 {
				continue
			}
			id := strings.ToLower(strings.TrimSpace(m[1]))
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true
			out = append(out, id)
		}
	}
	add(section)
	lowerStatus := strings.ToLower(status)
	if strings.Contains(lowerStatus, "supersedes") && !strings.Contains(lowerStatus, "superseded by") && !strings.Contains(lowerStatus, "superceded by") {
		add(status)
	}
	return out
}

func parseComponentList(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	out := make([]string, 0)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimLeft(line, "-*"))
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func buildADRStatement(snap domain.ADRSnapshot) string {
	var b strings.Builder
	title := strings.TrimSpace(snap.Title)
	decision := strings.TrimSpace(snap.Decision)
	if title != "" {
		b.WriteString(title)
		if decision != "" && !strings.EqualFold(title, decision) {
			b.WriteString("\n\n")
			b.WriteString(decision)
		}
	} else {
		b.WriteString(decision)
	}
	appendSection := func(label, body string) {
		body = strings.TrimSpace(body)
		if body == "" {
			return
		}
		b.WriteString("\n\n")
		b.WriteString(label)
		b.WriteString("\n")
		b.WriteString(body)
	}
	appendSection("Context:", snap.Context)
	appendSection("Alternatives:", snap.Alternatives)
	appendSection("Consequences:", snap.Consequences)
	return strings.TrimSpace(b.String())
}
