package output

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aric/garden/internal/agents"
	"github.com/aric/garden/internal/memory"
	"github.com/aric/garden/internal/retrieval"
)

type AgentsChange struct {
	Path     string
	Before   string
	After    string
	Applied  bool
	Findings []agents.Finding
}

func WritePack(w io.Writer, path string, task string, results []retrieval.Result, explain bool) error {
	if _, err := fmt.Fprintln(w, "<garden_context_pack>"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "# Garden Context Pack"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Path: `%s`\n", path); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Task: %s\n", task); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "## Relevant Memories"); err != nil {
		return err
	}
	if len(results) == 0 {
		if _, err := fmt.Fprintln(w, "No relevant memories."); err != nil {
			return err
		}
	} else {
		for _, result := range results {
			if _, err := fmt.Fprintf(w, "- %s\n", oneLine(result.Memory.Memory)); err != nil {
				return err
			}
		}
	}
	if explain && len(results) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "## Why These Memories"); err != nil {
			return err
		}
		for _, result := range results {
			if _, err := fmt.Fprintf(w, "%s selected:\n", result.Memory.ID); err != nil {
				return err
			}
			if len(result.Reasons) == 0 {
				if _, err := fmt.Fprintf(w, "- retrieval score (%s)\n", formatPoints(result.Score)); err != nil {
					return err
				}
				continue
			}
			for _, reason := range result.Reasons {
				if _, err := fmt.Fprintf(w, "- %s (%s)\n", reason.Text, formatPoints(reason.Points)); err != nil {
					return err
				}
			}
		}
	}
	_, err := fmt.Fprintln(w, "</garden_context_pack>")
	return err
}

func WriteList(w io.Writer, memories []memory.Memory) error {
	if len(memories) == 0 {
		_, err := fmt.Fprintln(w, "No memories.")
		return err
	}
	for _, mem := range memories {
		parts := []string{fmt.Sprintf("%s [%s]", mem.ID, mem.Priority)}
		if mem.Always {
			parts = append(parts, "always")
		}
		if len(mem.Scope) > 0 {
			parts = append(parts, "scope: "+strings.Join(mem.Scope, ", "))
		}
		if len(mem.Tags) > 0 {
			parts = append(parts, "tags: "+strings.Join(mem.Tags, ", "))
		}
		if _, err := fmt.Fprintf(w, "%s\n  %s\n", strings.Join(parts, " | "), oneLine(mem.Memory)); err != nil {
			return err
		}
	}
	return nil
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

func WriteAgentsLint(w io.Writer, findings []agents.Finding) (bool, error) {
	if len(findings) == 0 {
		_, err := fmt.Fprintln(w, "AGENTS.md lint passed.")
		return false, err
	}
	_, err := writeFindings(w, findings)
	return hasErrorFinding(findings), err
}

func oneLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func formatPoints(points int) string {
	if points > 0 {
		return fmt.Sprintf("+%d", points)
	}
	return fmt.Sprintf("%d", points)
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
