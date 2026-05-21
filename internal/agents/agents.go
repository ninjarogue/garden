package agents

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

const (
	AgentsStartMarker = "<!-- garden:agents:start -->"
	AgentsEndMarker   = "<!-- garden:agents:end -->"
	IndexStartMarker  = "<!-- garden:index:start -->"
	IndexEndMarker    = "<!-- garden:index:end -->"
)

type Context struct {
	Purpose     string
	Setup       []string
	Build       []string
	Lint        []string
	Typecheck   []string
	Test        []string
	Structure   []Entry
	Conventions []string
	Docs        []Entry
	Notes       []string
}

type Entry struct {
	Path string
	Text string
}

type IndexMemory struct {
	ID     string
	Scope  []string
	Always bool
	Tags   []string
}

type LintOptions struct {
	MaxLines     int
	MaxBytes     int
	MemoryBodies []string
}

type Finding struct {
	Severity string
	Code     string
	Message  string
}

func ParseEntry(value string) (Entry, error) {
	path, text, ok := strings.Cut(value, ":")
	path = strings.TrimSpace(path)
	text = strings.TrimSpace(text)
	if !ok || path == "" || text == "" {
		return Entry{}, fmt.Errorf("expected <path>: <description>")
	}
	return Entry{Path: path, Text: text}, nil
}

func RenderBlock(ctx Context, memories []IndexMemory) (string, error) {
	if err := validateContext(ctx); err != nil {
		return "", err
	}
	index, err := RenderIndex(memories)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(AgentsStartMarker)
	b.WriteString("\n")
	b.WriteString("## Garden Agent Context\n")

	if strings.TrimSpace(ctx.Purpose) != "" {
		b.WriteString("\n### Project Purpose\n")
		b.WriteString(strings.TrimSpace(ctx.Purpose))
		b.WriteString("\n")
	}

	writeCommands(&b, "Validation", []commandGroup{
		{Label: "Setup", Values: ctx.Setup},
		{Label: "Build", Values: ctx.Build},
		{Label: "Lint", Values: ctx.Lint},
		{Label: "Typecheck", Values: ctx.Typecheck},
		{Label: "Test", Values: ctx.Test},
	})
	writeEntries(&b, "Project Structure", ctx.Structure)
	writeList(&b, "Conventions", ctx.Conventions)
	writeEntries(&b, "Docs", ctx.Docs)
	writeList(&b, "Notes", ctx.Notes)

	b.WriteString("\n### Garden Memory\n")
	b.WriteString("This repo uses Garden for scoped repo memory.\n\n")
	b.WriteString("Source of truth: `.garden/memories.json`\n\n")
	b.WriteString("Before coding in an area with matching memory, run:\n\n")
	b.WriteString("`garden pack --path <file-or-dir> --task \"<what you are doing>\"`\n\n")
	b.WriteString("Memory index:\n")
	b.WriteString(IndexStartMarker)
	b.WriteString("\n")
	b.WriteString(index)
	b.WriteString(IndexEndMarker)
	b.WriteString("\n")
	b.WriteString(AgentsEndMarker)
	b.WriteString("\n")
	return b.String(), nil
}

func RenderIndex(memories []IndexMemory) (string, error) {
	if err := validateIndexMemories(memories); err != nil {
		return "", err
	}

	rows := map[string]*indexRow{}
	for _, mem := range memories {
		scopes := compactScopes(mem)
		for _, scope := range scopes {
			row := rows[scope]
			if row == nil {
				row = &indexRow{tags: map[string]bool{}, ids: map[string]bool{}}
				rows[scope] = row
			}
			for _, tag := range mem.Tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					row.tags[tag] = true
				}
			}
			if strings.TrimSpace(mem.ID) != "" {
				row.ids[strings.TrimSpace(mem.ID)] = true
			}
		}
	}

	var b strings.Builder
	b.WriteString("[Garden Memory Index]|root:.garden/memories.json\n")
	b.WriteString("|IMPORTANT:Prefer Garden repo memory over guessing when relevant\n")

	scopes := make([]string, 0, len(rows))
	for scope := range rows {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)
	for _, scope := range scopes {
		items := sortedKeys(rows[scope].tags)
		items = append(items, sortedKeys(rows[scope].ids)...)
		if len(items) == 0 {
			continue
		}
		b.WriteString("|")
		b.WriteString(scope)
		b.WriteString(":{")
		b.WriteString(strings.Join(items, ","))
		b.WriteString("}\n")
	}
	return b.String(), nil
}

func UpsertBlock(doc string, block string) (string, error) {
	rangeInfo, err := findSingleMarkerRange(doc, AgentsStartMarker, AgentsEndMarker, "Garden agents")
	if err != nil {
		return "", err
	}
	block = ensureTrailingNewline(block)
	if !rangeInfo.exists {
		return appendBlock(doc, block), nil
	}
	return doc[:rangeInfo.start] + block + doc[rangeInfo.end:], nil
}

func SyncIndex(doc string, memories []IndexMemory) (string, error) {
	agentsRange, err := findSingleMarkerRange(doc, AgentsStartMarker, AgentsEndMarker, "Garden agents")
	if err != nil {
		return "", err
	}
	if !agentsRange.exists {
		return "", fmt.Errorf("Garden agents block is missing; run garden agents update first")
	}
	index, err := RenderIndex(memories)
	if err != nil {
		return "", err
	}

	block := doc[agentsRange.start:agentsRange.end]
	newBlock, err := syncIndexInBlock(block, index)
	if err != nil {
		return "", err
	}
	return doc[:agentsRange.start] + newBlock + doc[agentsRange.end:], nil
}

func Lint(doc string, opts LintOptions) []Finding {
	findings := []Finding{}
	if opts.MaxLines > 0 && lineCount(doc) > opts.MaxLines {
		findings = append(findings, warning("line-budget", fmt.Sprintf("AGENTS.md has %d lines, over budget %d", lineCount(doc), opts.MaxLines)))
	}
	if opts.MaxBytes > 0 && len([]byte(doc)) > opts.MaxBytes {
		findings = append(findings, warning("size-budget", fmt.Sprintf("AGENTS.md has %d bytes, over budget %d", len([]byte(doc)), opts.MaxBytes)))
	}

	rangeInfo, err := findSingleMarkerRange(doc, AgentsStartMarker, AgentsEndMarker, "Garden agents")
	if err != nil {
		findings = append(findings, Finding{Severity: "error", Code: "garden-agents-markers", Message: err.Error()})
		return findings
	}
	if !rangeInfo.exists {
		findings = append(findings, warning("missing-garden-agents-block", "Garden agents block is missing"))
		return findings
	}

	block := doc[rangeInfo.start:rangeInfo.end]
	if !strings.Contains(block, "### Project Purpose") {
		findings = append(findings, warning("missing-project-purpose", "Garden agents block is missing Project Purpose"))
	}
	if !strings.Contains(block, "### Validation") {
		findings = append(findings, warning("missing-validation", "Garden agents block is missing Validation"))
	}
	if !strings.Contains(block, "### Project Structure") {
		findings = append(findings, warning("missing-project-structure", "Garden agents block is missing Project Structure"))
	}
	if _, err := findSingleMarkerRange(block, IndexStartMarker, IndexEndMarker, "Garden index"); err != nil {
		findings = append(findings, Finding{Severity: "error", Code: "garden-index-markers", Message: err.Error()})
	}

	for _, body := range opts.MemoryBodies {
		body = strings.TrimSpace(body)
		if body != "" && strings.Contains(doc, body) {
			findings = append(findings, warning("full-memory-body", "AGENTS.md appears to contain a full Garden memory body"))
			break
		}
	}
	for _, risky := range []string{"password", "secret", "api key", "apikey", "token"} {
		if strings.Contains(strings.ToLower(doc), risky) {
			findings = append(findings, warning("secret-like-content", "AGENTS.md contains secret-like wording; do not store secrets in agent instructions"))
			break
		}
	}

	return findings
}

type commandGroup struct {
	Label  string
	Values []string
}

type indexRow struct {
	tags map[string]bool
	ids  map[string]bool
}

type markerRange struct {
	exists bool
	start  int
	end    int
}

func writeCommands(b *strings.Builder, title string, groups []commandGroup) {
	if !hasAnyCommand(groups) {
		return
	}
	b.WriteString("\n### ")
	b.WriteString(title)
	b.WriteString("\n")
	for _, group := range groups {
		for _, value := range cleanStrings(group.Values) {
			b.WriteString("- ")
			b.WriteString(group.Label)
			b.WriteString(": `")
			b.WriteString(value)
			b.WriteString("`\n")
		}
	}
}

func writeEntries(b *strings.Builder, title string, entries []Entry) {
	cleaned := cleanEntries(entries)
	if len(cleaned) == 0 {
		return
	}
	b.WriteString("\n### ")
	b.WriteString(title)
	b.WriteString("\n")
	for _, entry := range cleaned {
		b.WriteString("- `")
		b.WriteString(entry.Path)
		b.WriteString("`: ")
		b.WriteString(entry.Text)
		b.WriteString("\n")
	}
}

func writeList(b *strings.Builder, title string, values []string) {
	values = cleanStrings(values)
	if len(values) == 0 {
		return
	}
	b.WriteString("\n### ")
	b.WriteString(title)
	b.WriteString("\n")
	for _, value := range values {
		b.WriteString("- ")
		b.WriteString(value)
		b.WriteString("\n")
	}
}

func hasAnyCommand(groups []commandGroup) bool {
	for _, group := range groups {
		if len(cleanStrings(group.Values)) > 0 {
			return true
		}
	}
	return false
}

func cleanEntries(entries []Entry) []Entry {
	cleaned := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		path := strings.TrimSpace(entry.Path)
		text := strings.TrimSpace(entry.Text)
		if path != "" && text != "" {
			cleaned = append(cleaned, Entry{Path: path, Text: text})
		}
	}
	return cleaned
}

func cleanStrings(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}
	return cleaned
}

func compactScopes(mem IndexMemory) []string {
	if mem.Always && len(mem.Scope) == 0 {
		return []string{"**/*"}
	}
	scopes := cleanStrings(mem.Scope)
	return scopes
}

func validateContext(ctx Context) error {
	if err := rejectReservedMarker("--purpose", ctx.Purpose); err != nil {
		return err
	}
	for _, group := range []struct {
		name   string
		values []string
	}{
		{name: "--setup", values: ctx.Setup},
		{name: "--build", values: ctx.Build},
		{name: "--lint", values: ctx.Lint},
		{name: "--typecheck", values: ctx.Typecheck},
		{name: "--test", values: ctx.Test},
		{name: "--convention", values: ctx.Conventions},
		{name: "--note", values: ctx.Notes},
	} {
		for _, value := range group.values {
			if err := rejectReservedMarker(group.name, value); err != nil {
				return err
			}
		}
	}
	for _, group := range []struct {
		name    string
		entries []Entry
	}{
		{name: "--map", entries: ctx.Structure},
		{name: "--doc", entries: ctx.Docs},
	} {
		for _, entry := range group.entries {
			if err := rejectReservedMarker(group.name, entry.Path); err != nil {
				return err
			}
			if err := rejectReservedMarker(group.name, entry.Text); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateIndexMemories(memories []IndexMemory) error {
	for _, mem := range memories {
		if err := rejectReservedMarker("memory id", mem.ID); err != nil {
			return err
		}
		if err := rejectCompactIndexItemSyntax("memory id", mem.ID); err != nil {
			return err
		}
		for _, scope := range mem.Scope {
			if err := rejectReservedMarker("memory scope", scope); err != nil {
				return err
			}
			if err := rejectCompactIndexRowSyntax("memory scope", scope); err != nil {
				return err
			}
		}
		for _, tag := range mem.Tags {
			if err := rejectReservedMarker("memory tag", tag); err != nil {
				return err
			}
			if err := rejectCompactIndexItemSyntax("memory tag", tag); err != nil {
				return err
			}
		}
	}
	return nil
}

func rejectReservedMarker(field string, value string) error {
	for _, marker := range []string{AgentsStartMarker, AgentsEndMarker, IndexStartMarker, IndexEndMarker} {
		if strings.Contains(value, marker) {
			return fmt.Errorf("%s contains reserved Garden marker %s", field, marker)
		}
	}
	return nil
}

func rejectCompactIndexRowSyntax(field string, value string) error {
	for _, r := range value {
		switch r {
		case '|':
			return fmt.Errorf("%s contains compact index syntax delimiter %q", field, r)
		}
		if unicode.IsControl(r) {
			return fmt.Errorf("%s contains compact index syntax control character", field)
		}
	}
	return nil
}

func rejectCompactIndexItemSyntax(field string, value string) error {
	for _, r := range value {
		switch r {
		case '|', '{', '}', ',':
			return fmt.Errorf("%s contains compact index syntax delimiter %q", field, r)
		}
		if unicode.IsControl(r) {
			return fmt.Errorf("%s contains compact index syntax control character", field)
		}
	}
	return nil
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func findSingleMarkerRange(doc string, startMarker string, endMarker string, label string) (markerRange, error) {
	startCount := strings.Count(doc, startMarker)
	endCount := strings.Count(doc, endMarker)
	if startCount == 0 && endCount == 0 {
		return markerRange{}, nil
	}
	if startCount != 1 || endCount != 1 {
		return markerRange{}, fmt.Errorf("malformed %s markers", label)
	}
	start := strings.Index(doc, startMarker)
	endStart := strings.Index(doc, endMarker)
	if start < 0 || endStart < 0 || start > endStart {
		return markerRange{}, fmt.Errorf("malformed %s markers", label)
	}
	return markerRange{exists: true, start: start, end: endStart + len(endMarker)}, nil
}

func appendBlock(doc string, block string) string {
	if doc == "" {
		return block
	}
	if strings.HasSuffix(doc, "\n\n") {
		return doc + block
	}
	if strings.HasSuffix(doc, "\n") {
		return doc + "\n" + block
	}
	return doc + "\n\n" + block
}

func ensureTrailingNewline(value string) string {
	if strings.HasSuffix(value, "\n") {
		return value
	}
	return value + "\n"
}

func syncIndexInBlock(block string, index string) (string, error) {
	index = IndexStartMarker + "\n" + ensureTrailingNewline(index) + IndexEndMarker + "\n"
	rangeInfo, err := findSingleMarkerRange(block, IndexStartMarker, IndexEndMarker, "Garden index")
	if err != nil {
		return "", err
	}
	if rangeInfo.exists {
		return block[:rangeInfo.start] + index + block[rangeInfo.end:], nil
	}

	end := strings.Index(block, AgentsEndMarker)
	if end < 0 {
		return "", fmt.Errorf("malformed Garden agents markers")
	}
	prefix := strings.TrimRight(block[:end], "\n") + "\n\n"
	return prefix + index + block[end:], nil
}

func lineCount(value string) int {
	if value == "" {
		return 0
	}
	count := strings.Count(value, "\n")
	if !strings.HasSuffix(value, "\n") {
		count++
	}
	return count
}

func warning(code string, message string) Finding {
	return Finding{Severity: "warning", Code: code, Message: message}
}
