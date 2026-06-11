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
	}, "sync", false); err != nil {
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
	}, "sync", false); err != nil {
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

func TestWriteAgentsChangeAppliedQuietWritesOnlyAppliedMessage(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteAgentsChange(&buf, app.AgentsChange{
		Path:    "AGENTS.md",
		Before:  "old\n",
		After:   "new\n",
		Applied: true,
	}, "sync", false); err != nil {
		t.Fatalf("WriteAgentsChange returned error: %v", err)
	}

	want := "Applied AGENTS.md sync.\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestWriteAgentsChangePreviewNoChangesWritesNoChangesAndPreviewMessage(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteAgentsChange(&buf, app.AgentsChange{
		Path:    "AGENTS.md",
		Before:  "same\n",
		After:   "same\n",
		Applied: false,
	}, "sync", false); err != nil {
		t.Fatalf("WriteAgentsChange returned error: %v", err)
	}

	want := "No changes for AGENTS.md\nPreview only. Re-run with --apply to write AGENTS.md.\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestWriteAgentsChangeAppliedVerboseWritesDiffAndAppliedMessage(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteAgentsChange(&buf, app.AgentsChange{
		Path:    "AGENTS.md",
		Before:  "old\n",
		After:   "new\n",
		Applied: true,
	}, "sync", true); err != nil {
		t.Fatalf("WriteAgentsChange returned error: %v", err)
	}

	want := "--- AGENTS.md\n+++ AGENTS.md\n@@\n-old\n+new\nApplied AGENTS.md sync.\n"
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

func TestWriteCheckReportWritesCompactReviewContext(t *testing.T) {
	var buf bytes.Buffer
	err := WriteCheckReport(&buf, app.CheckReport{ChangedFiles: []app.CheckChangedFile{{
		Path: "internal/cmd/root.go",
		Cards: []app.CheckMatchedCard{{
			Path:         ".garden/context/app-layer-architecture.md",
			MatchedScope: "internal/cmd/**",
		}},
	}}, SuggestedVerifications: []app.CheckSuggestedVerification{{
		Path:         ".garden/context/app-layer-architecture.md",
		Verification: "Run:\n\n```sh\nenv GOCACHE=/tmp/garden-go-build go test ./...\n```",
	}}})
	if err != nil {
		t.Fatalf("WriteCheckReport returned error: %v", err)
	}

	want := "Garden review context\n" +
		"\n" +
		"Changed:\n" +
		"  internal/cmd/root.go\n" +
		"\n" +
		"Relevant constraints:\n" +
		"  internal/cmd/root.go\n" +
		"    .garden/context/app-layer-architecture.md\n" +
		"    matched: internal/cmd/**\n" +
		"\n" +
		"Suggested verification:\n" +
		"  .garden/context/app-layer-architecture.md\n" +
		"    Run:\n" +
		"\n" +
		"    ```sh\n" +
		"    env GOCACHE=/tmp/garden-go-build go test ./...\n" +
		"    ```\n" +
		"\n" +
		"Verification surfaces changed:\n" +
		"  none\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestWriteCheckReportWritesNoMatchesAndWarnings(t *testing.T) {
	var buf bytes.Buffer
	err := WriteCheckReport(&buf, app.CheckReport{
		ChangedFiles: []app.CheckChangedFile{{Path: "README.md"}},
		Warnings: []app.CheckWarning{{
			Path:    "internal/cmd/root_test.go",
			Code:    "verification-surface-changed",
			Message: "changed test file",
		}},
	})
	if err != nil {
		t.Fatalf("WriteCheckReport returned error: %v", err)
	}

	want := "Garden review context\n" +
		"\n" +
		"Changed:\n" +
		"  README.md\n" +
		"\n" +
		"Relevant constraints:\n" +
		"  README.md\n" +
		"    none\n" +
		"\n" +
		"Suggested verification:\n" +
		"  none\n" +
		"\n" +
		"Verification surfaces changed:\n" +
		"  internal/cmd/root_test.go: changed test file\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}
