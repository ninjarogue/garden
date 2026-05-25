package output

import (
	"bytes"
	"testing"

	"github.com/aric/garden/internal/app"
)

func TestWriteAgentsChangePreviewWritesDiffAndPreviewMessage(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteAgentsChange(&buf, app.AgentsChange{
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
	if err := WriteAgentsChange(&buf, app.AgentsChange{
		Path:    "AGENTS.md",
		Before:  "old\n",
		After:   "new\n",
		Applied: false,
		Findings: []app.Finding{{
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
	if err := WriteAgentsChange(&buf, app.AgentsChange{
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
	failed, err := WriteLint(&buf, []app.Finding{
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

func TestWriteCardsWritesCardMetadata(t *testing.T) {
	var buf bytes.Buffer
	err := WriteCards(&buf, []app.Card{{
		Path:  ".garden/context/routes-query-modules.md",
		Scope: []string{"src/routes/**", "src/db/**"},
		Tags:  []string{"database", "tenant-scoping"},
	}})
	if err != nil {
		t.Fatalf("WriteCards returned error: %v", err)
	}

	want := `.garden/context/routes-query-modules.md
  scope: src/routes/**, src/db/**
  tags: database, tenant-scoping
`
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestWriteCardsWritesEmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCards(&buf, nil); err != nil {
		t.Fatalf("WriteCards returned error: %v", err)
	}

	if buf.String() != "No context cards found.\n" {
		t.Fatalf("output = %q", buf.String())
	}
}
