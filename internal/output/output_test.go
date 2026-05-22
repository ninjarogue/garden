package output

import (
	"bytes"
	"testing"

	"github.com/aric/garden/internal/agents"
)

func TestWriteAgentsChangePreviewWritesDiffAndPreviewMessage(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteAgentsChange(&buf, AgentsChange{
		Path:    "AGENTS.md",
		Before:  "old\n",
		After:   "new\n",
		Applied: false,
	}, "sync"); err != nil {
		t.Fatalf("WriteAgentsChange returned error: %v", err)
	}

	want := "--- AGENTS.md\n+++ AGENTS.md\n@@\n-old\n+new\nPreview only. Re-run with --apply to write AGENTS.md.\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestWriteAgentsChangePreviewIncludesLintFindings(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteAgentsChange(&buf, AgentsChange{
		Path:    "AGENTS.md",
		Before:  "old\n",
		After:   "new\n",
		Applied: false,
		Findings: []agents.Finding{{
			Severity: "error",
			Code:     "stale-garden-index",
			Message:  "AGENTS.md Garden index is stale",
		}},
	}, "sync"); err != nil {
		t.Fatalf("WriteAgentsChange returned error: %v", err)
	}

	want := "--- AGENTS.md\n+++ AGENTS.md\n@@\n-old\n+new\n" +
		"Lint findings:\n" +
		"error stale-garden-index: AGENTS.md Garden index is stale\n" +
		"Preview only. Re-run with --apply to write AGENTS.md.\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestWriteAgentsChangeAppliedWritesNoChangesAndAppliedMessage(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteAgentsChange(&buf, AgentsChange{
		Path:    "AGENTS.md",
		Before:  "same\n",
		After:   "same\n",
		Applied: true,
	}, "sync"); err != nil {
		t.Fatalf("WriteAgentsChange returned error: %v", err)
	}

	want := "No changes for AGENTS.md\nApplied AGENTS.md sync.\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestWriteLintPassWritesPassedMessage(t *testing.T) {
	var buf bytes.Buffer
	failed, err := WriteLint(&buf, nil)
	if err != nil {
		t.Fatalf("WriteLint returned error: %v", err)
	}
	if failed {
		t.Fatal("expected nil findings to pass")
	}

	want := "Garden lint passed.\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestWriteLintFindingsWritesFindingsAndFailsOnErrors(t *testing.T) {
	var buf bytes.Buffer
	failed, err := WriteLint(&buf, []agents.Finding{
		{Severity: "warning", Code: "line-budget", Message: "too long"},
		{Severity: "error", Code: "stale-garden-index", Message: "stale"},
	})
	if err != nil {
		t.Fatalf("WriteLint returned error: %v", err)
	}
	if !failed {
		t.Fatal("expected error finding to mark lint failed")
	}

	want := "warning line-budget: too long\nerror stale-garden-index: stale\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}
