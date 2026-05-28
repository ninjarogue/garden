package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aric/garden/internal/agents"
	"github.com/aric/garden/internal/app"
)

func TestInitCommandCreatesContextDirectory(t *testing.T) {
	rootDir := t.TempDir()
	garden := app.New(app.Options{Root: rootDir})

	out, _, err := execute(garden, "init")
	if err != nil {
		t.Fatalf("init returned error: %v", err)
	}
	if out != "Initialized .garden/context\n" {
		t.Fatalf("stdout = %q", out)
	}
	if _, err := os.Stat(filepath.Join(rootDir, ".garden", "context")); err != nil {
		t.Fatalf("expected context directory: %v", err)
	}
}

func TestNewCommandCreatesMarkdownContextCard(t *testing.T) {
	rootDir := t.TempDir()
	garden := app.New(app.Options{Root: rootDir})

	out, _, err := execute(garden,
		"new",
		"routes-query-modules",
		"--scope", "src/routes/**",
		"--tag", "database",
		"--tag", "tenant-scoping",
	)
	if err != nil {
		t.Fatalf("new returned error: %v", err)
	}
	if out != "Created .garden/context/routes-query-modules.md\n" {
		t.Fatalf("stdout = %q", out)
	}

	cardData, err := os.ReadFile(filepath.Join(rootDir, ".garden", "context", "routes-query-modules.md"))
	if err != nil {
		t.Fatalf("read card: %v", err)
	}
	wantCard := `---
scope:
  - src/routes/**
tags:
  - database
  - tenant-scoping
---

# Routes Query Modules

Write the repo context here.
`
	if string(cardData) != wantCard {
		t.Fatalf("card content = %q, want %q", string(cardData), wantCard)
	}
}

func TestAgentsSyncCommandAppliesContextIndex(t *testing.T) {
	rootDir := t.TempDir()
	garden := app.New(app.Options{Root: rootDir})
	writeCard(t, rootDir, "routes-query-modules", `---
scope:
  - src/routes/**
tags:
  - database
  - tenant-scoping
---

# Routes Query Modules

Use query modules.
`)

	out, _, err := execute(garden, "agents", "sync", "--apply")
	if err != nil {
		t.Fatalf("agents sync returned error: %v", err)
	}
	assertContains(t, out, "--- AGENTS.md\n+++ AGENTS.md\n@@\n")
	assertContains(t, out, "|src/routes/**:.garden/context/routes-query-modules.md")
	assertContains(t, out, "Applied AGENTS.md sync.\n")

	agentsData, err := os.ReadFile(filepath.Join(rootDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	assertContains(t, string(agentsData), agents.AgentsStartMarker)
	assertContains(t, string(agentsData), "|src/routes/**:.garden/context/routes-query-modules.md")
}

func TestLintCommandPassesWhenAgentsIndexIsCurrent(t *testing.T) {
	rootDir := t.TempDir()
	garden := app.New(app.Options{Root: rootDir})
	writeCard(t, rootDir, "routes-query-modules", `---
scope:
  - src/routes/**
tags:
  - database
---

Use query modules.
`)
	block, err := agents.RenderBlock([]agents.IndexCard{{
		Path:  ".garden/context/routes-query-modules.md",
		Scope: []string{"src/routes/**"},
	}})
	if err != nil {
		t.Fatalf("RenderBlock returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "AGENTS.md"), []byte(block), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	out, _, err := execute(garden, "lint")
	if err != nil {
		t.Fatalf("lint returned error: %v", err)
	}
	if out != "Garden lint passed.\n" {
		t.Fatalf("stdout = %q", out)
	}
}

func TestLintCommandReturnsErrorWhenLintFindsProblems(t *testing.T) {
	rootDir := t.TempDir()
	garden := app.New(app.Options{Root: rootDir})
	writeCard(t, rootDir, "routes-query-modules", `---
scope:
  - src/routes/**
---

Use query modules.
`)
	staleAgents := agents.AgentsStartMarker + "\n### Garden Context\n" +
		agents.IndexStartMarker + "\n" +
		"[Garden Context Index]|root:.garden/context\n" +
		"|old/**:{old,.garden/context/old.md}\n" +
		agents.IndexEndMarker + "\n" +
		agents.AgentsEndMarker + "\n"
	if err := os.WriteFile(filepath.Join(rootDir, "AGENTS.md"), []byte(staleAgents), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	out, _, err := execute(garden, "lint")
	if err == nil {
		t.Fatal("expected lint command error")
	}
	if err.Error() != "garden lint failed" {
		t.Fatalf("error = %q, want garden lint failed", err.Error())
	}
	wantOut := "error stale-garden-index: AGENTS.md Garden index is stale; run garden agents sync --apply\n"
	if out != wantOut {
		t.Fatalf("stdout = %q, want %q", out, wantOut)
	}
}

func TestLintCommandReturnsErrorForInvalidScopeGlob(t *testing.T) {
	rootDir := t.TempDir()
	garden := app.New(app.Options{Root: rootDir})
	writeCard(t, rootDir, "broken-scope", `---
scope:
  - internal/[*.go
---

Use query modules.
`)
	block, err := agents.RenderBlock(nil)
	if err != nil {
		t.Fatalf("RenderBlock returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootDir, "AGENTS.md"), []byte(block), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	out, _, err := execute(garden, "lint")
	if err == nil {
		t.Fatal("expected lint command error")
	}
	if err.Error() != "garden lint failed" {
		t.Fatalf("error = %q, want garden lint failed", err.Error())
	}
	assertContains(t, out, `error invalid-context-card: .garden/context/broken-scope.md: invalid scope glob "internal/[*.go"`)
}

func TestCheckCommandReportsReviewContextForChangedPaths(t *testing.T) {
	rootDir := t.TempDir()
	garden := app.New(app.Options{Root: rootDir})
	writeCard(t, rootDir, "app-layer-architecture", `---
scope:
  - internal/app/**
  - internal/cmd/**
---

# App Layer Architecture

Keep commands thin.

## Verification

Run tests.
`)

	out, _, err := execute(garden,
		"check",
		"internal/cmd/root.go",
		"internal/cmd/root_test.go",
	)
	if err != nil {
		t.Fatalf("check returned error: %v", err)
	}
	assertContains(t, out, "Garden review context\n")
	assertContains(t, out, "  internal/cmd/root.go\n")
	assertContains(t, out, "  internal/cmd/root_test.go\n")
	assertContains(t, out, "    .garden/context/app-layer-architecture.md\n")
	assertContains(t, out, "    matched: internal/cmd/**\n")
	assertContains(t, out, "    Run tests.\n")
	assertContains(t, out, "  internal/cmd/root_test.go: changed test file\n")
}

func TestCheckCommandRequiresChangedPath(t *testing.T) {
	garden := app.New(app.Options{Root: t.TempDir()})

	_, _, err := execute(garden, "check")
	if err == nil {
		t.Fatal("expected check command error")
	}
	if !strings.Contains(err.Error(), "at least one changed path is required") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestRemoveCommandDeletesContextCard(t *testing.T) {
	rootDir := t.TempDir()
	garden := app.New(app.Options{Root: rootDir})
	writeCard(t, rootDir, "routes-query-modules", `---
scope:
  - src/routes/**
---

Use query modules.
`)

	out, _, err := execute(garden, "remove", "routes-query-modules")
	if err != nil {
		t.Fatalf("remove returned error: %v", err)
	}
	if out != "Removed .garden/context/routes-query-modules.md\n" {
		t.Fatalf("stdout = %q", out)
	}
	if _, err := os.Stat(filepath.Join(rootDir, ".garden", "context", "routes-query-modules.md")); !os.IsNotExist(err) {
		t.Fatalf("expected card to be removed, stat err = %v", err)
	}
}

func TestAgentsSyncPreviewsByDefault(t *testing.T) {
	rootDir := t.TempDir()
	garden := app.New(app.Options{Root: rootDir})
	writeCard(t, rootDir, "routes-query-modules", `---
scope:
  - src/routes/**
---

Use query modules.
`)

	out, _, err := execute(garden, "agents", "sync")
	if err != nil {
		t.Fatalf("agents sync preview returned error: %v", err)
	}
	assertContains(t, out, "--- AGENTS.md\n+++ AGENTS.md\n@@\n")
	assertContains(t, out, "|src/routes/**:.garden/context/routes-query-modules.md")
	assertContains(t, out, "Preview only. Re-run with --apply to write AGENTS.md.\n")
	if _, err := os.Stat(filepath.Join(rootDir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("preview wrote AGENTS.md, stat err = %v", err)
	}
}

func TestCommandValidationReturnsActionableErrors(t *testing.T) {
	garden := app.New(app.Options{Root: t.TempDir()})

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "new requires exactly one slug", args: []string{"new", "one", "two", "--scope", "**/*"}, wantErr: "accepts exactly one context card slug"},
		{name: "remove requires exactly one slug", args: []string{"remove", "one", "two"}, wantErr: "accepts exactly one context card slug"},
		{name: "new requires scope", args: []string{"new", "routes-query-modules"}, wantErr: "scope must include at least one glob"},
		{name: "new rejects invalid slug", args: []string{"new", "Routes_Query_Modules", "--scope", "**/*"}, wantErr: "invalid card slug"},
		{name: "remember is not a core command", args: []string{"remember", "Use query modules.", "--scope", "**/*"}, wantErr: "unknown command \"remember\""},
		{name: "pack is not a core command", args: []string{"pack", "--path", "src/file.go", "--task", "add endpoint"}, wantErr: "unknown command \"pack\""},
		{name: "agents update is not a core command", args: []string{"agents", "update"}, wantErr: "unknown command \"update\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := execute(garden, tt.args...)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func execute(garden *app.App, args ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := NewRoot(Options{App: garden, Stdout: &stdout, Stderr: &stderr})
	root.SetArgs(args)
	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

func writeCard(t *testing.T, root string, slug string, content string) {
	t.Helper()
	dir := filepath.Join(root, ".garden", "context")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir context: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, slug+".md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write card: %v", err)
	}
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("value = %q, want substring %q", got, want)
	}
}
