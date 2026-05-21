package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aric/garden/internal/agents"
	"github.com/aric/garden/internal/memory"
	"github.com/aric/garden/internal/retrieval"
)

func TestWritePackExplainIncludesReasonsInsideContextPack(t *testing.T) {
	var buf bytes.Buffer
	results := []retrieval.Result{{
		Memory: memory.Memory{ID: "mem_1111111111", Memory: "Use query modules."},
		Reasons: []retrieval.Reason{{
			Text:   "scope matched `src/routes/**`",
			Points: 40,
		}},
	}}

	if err := WritePack(&buf, "src/routes/api/users.ts", "add endpoint", results, true); err != nil {
		t.Fatalf("WritePack returned error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"<garden_context_pack>",
		"## Relevant Memories",
		"- Use query modules.",
		"## Why These Memories",
		"mem_1111111111 selected:",
		"- scope matched `src/routes/**` (+40)",
		"</garden_context_pack>",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("pack output missing %q:\n%s", want, out)
		}
	}
}

func TestWriteAgentsChangeFormatsPreviewApplyAndFindings(t *testing.T) {
	var preview bytes.Buffer
	if err := WriteAgentsChange(&preview, AgentsChange{
		Path:    "AGENTS.md",
		Before:  "old\n",
		After:   "new\n",
		Applied: false,
		Findings: []agents.Finding{{
			Severity: "warning",
			Code:     "line-budget",
			Message:  "too long",
		}},
	}, "update"); err != nil {
		t.Fatalf("WriteAgentsChange preview returned error: %v", err)
	}
	for _, want := range []string{
		"--- AGENTS.md\n+++ AGENTS.md\n@@\n-old\n+new\n",
		"Lint findings:\nwarning line-budget: too long\n",
		"Preview only. Re-run with --apply to write AGENTS.md.\n",
	} {
		if !strings.Contains(preview.String(), want) {
			t.Fatalf("preview output missing %q:\n%s", want, preview.String())
		}
	}

	var applied bytes.Buffer
	if err := WriteAgentsChange(&applied, AgentsChange{Path: "AGENTS.md", Before: "same\n", After: "same\n", Applied: true}, "sync"); err != nil {
		t.Fatalf("WriteAgentsChange apply returned error: %v", err)
	}
	if !strings.Contains(applied.String(), "No changes for AGENTS.md\n") {
		t.Fatalf("apply output = %q", applied.String())
	}
	if !strings.Contains(applied.String(), "Applied AGENTS.md sync.\n") {
		t.Fatalf("apply output = %q", applied.String())
	}
}

func TestWriteAgentsLintFormatsPassWarningsAndErrors(t *testing.T) {
	var passed bytes.Buffer
	failed, err := WriteAgentsLint(&passed, nil)
	if err != nil {
		t.Fatalf("WriteAgentsLint pass returned error: %v", err)
	}
	if failed || passed.String() != "AGENTS.md lint passed.\n" {
		t.Fatalf("failed = %v output = %q", failed, passed.String())
	}

	var findings bytes.Buffer
	failed, err = WriteAgentsLint(&findings, []agents.Finding{
		{Severity: "warning", Code: "line-budget", Message: "too long"},
		{Severity: "error", Code: "garden-agents-markers", Message: "malformed"},
	})
	if err != nil {
		t.Fatalf("WriteAgentsLint findings returned error: %v", err)
	}
	if !failed {
		t.Fatal("expected error finding to mark lint failed")
	}
	for _, want := range []string{
		"warning line-budget: too long\n",
		"error garden-agents-markers: malformed\n",
	} {
		if !strings.Contains(findings.String(), want) {
			t.Fatalf("lint output missing %q:\n%s", want, findings.String())
		}
	}
}
