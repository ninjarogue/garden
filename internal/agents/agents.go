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

type IndexCard struct {
	Path  string
	Kind  string
	Scope []string
	Tags  []string
}

type LintOptions struct {
	ExpectedIndex string
}

type Finding struct {
	Severity string
	Code     string
	Message  string
}

type indexRow struct {
	items []string
	seen  map[string]bool
}

type markerRange struct {
	exists bool
	start  int
	end    int
}

func RenderBlock(cards []IndexCard) (string, error) {
	index, err := RenderIndex(cards)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(AgentsStartMarker)
	b.WriteString("\n")
	b.WriteString("### Garden Context\n\n")
	b.WriteString("Detailed agent context lives in `.garden/context/*.md`.\n\n")
	b.WriteString("Before editing a listed area, inspect the matching context card.\n\n")
	b.WriteString("Index:\n")
	b.WriteString(IndexStartMarker)
	b.WriteString("\n")
	b.WriteString(index)
	b.WriteString(IndexEndMarker)
	b.WriteString("\n")
	b.WriteString(AgentsEndMarker)
	b.WriteString("\n")
	return b.String(), nil
}

func RenderIndex(cards []IndexCard) (string, error) {
	if err := validateIndexCards(cards); err != nil {
		return "", err
	}

	cards = sortedCards(cards)
	rows := map[string]*indexRow{}
	for _, card := range cards {
		for _, scope := range cleanStrings(card.Scope) {
			row := rows[scope]
			if row == nil {
				row = &indexRow{seen: map[string]bool{}}
				rows[scope] = row
			}
			row.add(strings.TrimSpace(card.Kind))
			for _, tag := range cleanStrings(card.Tags) {
				row.add(tag)
			}
			row.add(strings.TrimSpace(card.Path))
		}
	}

	scopes := make([]string, 0, len(rows))
	for scope := range rows {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)

	var b strings.Builder
	b.WriteString("[Garden Context Index]|root:.garden/context\n")
	b.WriteString("|IMPORTANT:Before editing a listed area, inspect the matching context card\n")
	for _, scope := range scopes {
		if len(rows[scope].items) == 0 {
			continue
		}
		b.WriteString("|")
		b.WriteString(scope)
		b.WriteString(":{")
		b.WriteString(strings.Join(rows[scope].items, ","))
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

func SyncIndex(doc string, cards []IndexCard) (string, error) {
	agentsRange, err := findSingleMarkerRange(doc, AgentsStartMarker, AgentsEndMarker, "Garden agents")
	if err != nil {
		return "", err
	}
	if !agentsRange.exists {
		block, err := RenderBlock(cards)
		if err != nil {
			return "", err
		}
		return UpsertBlock(doc, block)
	}

	index, err := RenderIndex(cards)
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
	rangeInfo, err := findSingleMarkerRange(doc, AgentsStartMarker, AgentsEndMarker, "Garden agents")
	if err != nil {
		return []Finding{errorFinding("garden-agents-markers", err.Error())}
	}
	if !rangeInfo.exists {
		return []Finding{errorFinding("missing-garden-agents-block", "Garden agents block is missing")}
	}

	block := doc[rangeInfo.start:rangeInfo.end]
	indexRange, err := findSingleMarkerRange(block, IndexStartMarker, IndexEndMarker, "Garden index")
	if err != nil {
		return []Finding{errorFinding("garden-index-markers", err.Error())}
	}
	if !indexRange.exists {
		return []Finding{errorFinding("missing-garden-index", "Garden context index is missing")}
	}
	if strings.TrimSpace(opts.ExpectedIndex) != "" {
		current := strings.TrimSpace(block[indexRange.start+len(IndexStartMarker) : indexRange.end-len(IndexEndMarker)])
		expected := strings.TrimSpace(opts.ExpectedIndex)
		if current != expected {
			return []Finding{errorFinding("stale-garden-index", "AGENTS.md Garden index is stale; run garden agents sync --apply")}
		}
	}
	return nil
}

func (r *indexRow) add(value string) {
	value = strings.TrimSpace(value)
	if value == "" || r.seen[value] {
		return
	}
	r.seen[value] = true
	r.items = append(r.items, value)
}

func sortedCards(cards []IndexCard) []IndexCard {
	sorted := append([]IndexCard(nil), cards...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})
	return sorted
}

func validateIndexCards(cards []IndexCard) error {
	for _, card := range cards {
		if err := rejectReservedMarker("context card path", card.Path); err != nil {
			return err
		}
		if err := rejectCompactIndexItemSyntax("context card path", card.Path); err != nil {
			return err
		}
		if err := rejectReservedMarker("context card kind", card.Kind); err != nil {
			return err
		}
		if err := rejectCompactIndexItemSyntax("context card kind", card.Kind); err != nil {
			return err
		}
		for _, scope := range card.Scope {
			if err := rejectReservedMarker("context card scope", scope); err != nil {
				return err
			}
			if err := rejectCompactIndexRowSyntax("context card scope", scope); err != nil {
				return err
			}
		}
		for _, tag := range card.Tags {
			if err := rejectReservedMarker("context card tag", tag); err != nil {
				return err
			}
			if err := rejectCompactIndexItemSyntax("context card tag", tag); err != nil {
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
	index = IndexStartMarker + "\n" + ensureTrailingNewline(index) + IndexEndMarker
	rangeInfo, err := findSingleMarkerRange(block, IndexStartMarker, IndexEndMarker, "Garden index")
	if err != nil {
		return "", err
	}
	if rangeInfo.exists {
		if !strings.HasPrefix(block[rangeInfo.end:], "\n") {
			index += "\n"
		}
		return block[:rangeInfo.start] + index + block[rangeInfo.end:], nil
	}

	end := strings.Index(block, AgentsEndMarker)
	if end < 0 {
		return "", fmt.Errorf("malformed Garden agents markers")
	}
	prefix := strings.TrimRight(block[:end], "\n") + "\n\n"
	return prefix + index + "\n" + block[end:], nil
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

func errorFinding(code string, message string) Finding {
	return Finding{Severity: "error", Code: code, Message: message}
}
