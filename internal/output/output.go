package output

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aric/garden/internal/app"
)

func WriteAgentsChange(w io.Writer, change app.AgentsChange, action string) error {
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

func WriteLint(w io.Writer, findings []app.Finding) (bool, error) {
	if len(findings) == 0 {
		_, err := fmt.Fprintln(w, "Garden lint passed.")
		return false, err
	}
	_, err := writeFindings(w, findings)
	return hasErrorFinding(findings), err
}

func WriteCheckReport(w io.Writer, report app.CheckReport) error {
	if _, err := fmt.Fprintln(w, "Garden review context"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Changed:"); err != nil {
		return err
	}
	for _, changedFile := range report.ChangedFiles {
		if _, err := fmt.Fprintf(w, "  %s\n", changedFile.Path); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Relevant constraints:"); err != nil {
		return err
	}
	for _, changedFile := range report.ChangedFiles {
		if _, err := fmt.Fprintf(w, "  %s\n", changedFile.Path); err != nil {
			return err
		}
		if len(changedFile.Cards) == 0 {
			if _, err := fmt.Fprintln(w, "    none"); err != nil {
				return err
			}
			continue
		}
		for _, card := range changedFile.Cards {
			if _, err := fmt.Fprintf(w, "    %s\n", card.Path); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "    matched: %s\n", card.MatchedScope); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Suggested verification:"); err != nil {
		return err
	}
	if err := writeSuggestedVerification(w, report); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Verification surfaces changed:"); err != nil {
		return err
	}
	if len(report.Warnings) == 0 {
		_, err := fmt.Fprintln(w, "  none")
		return err
	}
	for _, warning := range report.Warnings {
		if _, err := fmt.Fprintf(w, "  %s: %s\n", warning.Path, warning.Message); err != nil {
			return err
		}
	}
	return nil
}

func writeSuggestedVerification(w io.Writer, report app.CheckReport) error {
	wrote := false
	seen := map[string]bool{}
	for _, changedFile := range report.ChangedFiles {
		for _, card := range changedFile.Cards {
			if card.Verification == "" || seen[card.Path] {
				continue
			}
			seen[card.Path] = true
			wrote = true
			if _, err := fmt.Fprintf(w, "  %s\n", card.Path); err != nil {
				return err
			}
			if err := writeIndentedBlock(w, card.Verification, "    "); err != nil {
				return err
			}
		}
	}
	if !wrote {
		_, err := fmt.Fprintln(w, "  none")
		return err
	}
	return nil
}

func writeIndentedBlock(w io.Writer, content string, indent string) error {
	for _, line := range strings.Split(content, "\n") {
		if line == "" {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "%s%s\n", indent, line); err != nil {
			return err
		}
	}
	return nil
}

func writeFindings(w io.Writer, findings []app.Finding) (int, error) {
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

func hasErrorFinding(findings []app.Finding) bool {
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
