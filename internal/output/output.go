package output

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aric/garden/internal/agents"
)

type AgentsChange struct {
	Path     string
	Before   string
	After    string
	Applied  bool
	Findings []agents.Finding
}

func WriteAgentsChange(w io.Writer, change AgentsChange, action string) error {
	diff := unifiedDiff(filepath.Base(change.Path), change.Before, change.After)
	if _, err := fmt.Fprint(w, diff); err != nil {
		return err
	}
	if !strings.HasSuffix(diff, "\n") {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	if len(change.Findings) > 0 {
		if _, err := fmt.Fprintln(w, "Lint findings:"); err != nil {
			return err
		}
		if _, err := writeFindings(w, change.Findings); err != nil {
			return err
		}
	}
	if change.Applied {
		_, err := fmt.Fprintf(w, "Applied AGENTS.md %s.\n", action)
		return err
	}
	_, err := fmt.Fprintln(w, "Preview only. Re-run with --apply to write AGENTS.md.")
	return err
}

func WriteLint(w io.Writer, findings []agents.Finding) (bool, error) {
	if len(findings) == 0 {
		_, err := fmt.Fprintln(w, "Garden lint passed.")
		return false, err
	}
	_, err := writeFindings(w, findings)
	return hasErrorFinding(findings), err
}

func writeFindings(w io.Writer, findings []agents.Finding) (int, error) {
	written := 0
	for _, finding := range findings {
		n, err := fmt.Fprintf(w, "%s %s: %s\n", finding.Severity, finding.Code, finding.Message)
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func hasErrorFinding(findings []agents.Finding) bool {
	for _, finding := range findings {
		if finding.Severity == "error" {
			return true
		}
	}
	return false
}

func unifiedDiff(path string, oldContent string, newContent string) string {
	if oldContent == newContent {
		return fmt.Sprintf("No changes for %s\n", path)
	}
	var b strings.Builder
	b.WriteString("--- ")
	b.WriteString(path)
	b.WriteString("\n+++ ")
	b.WriteString(path)
	b.WriteString("\n@@\n")
	for _, line := range diffLines(oldContent) {
		b.WriteString("-")
		b.WriteString(line)
		b.WriteString("\n")
	}
	for _, line := range diffLines(newContent) {
		b.WriteString("+")
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}

func diffLines(content string) []string {
	content = strings.TrimSuffix(content, "\n")
	if content == "" {
		return []string{}
	}
	return strings.Split(content, "\n")
}
