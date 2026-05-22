package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aric/garden/internal/agents"
)

func TestNewCardCreatesMarkdownCard(t *testing.T) {
	root := t.TempDir()
	garden := New(Options{Root: root})
	if err := garden.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	card, err := garden.NewCard(NewCardInput{
		Slug:  "routes-query-modules",
		Scope: []string{"src/routes/**"},
		Tags:  []string{"database"},
	})
	if err != nil {
		t.Fatalf("NewCard returned error: %v", err)
	}
	if card.Path != ".garden/context/routes-query-modules.md" {
		t.Fatalf("Path = %q", card.Path)
	}

	data, err := os.ReadFile(filepath.Join(root, ".garden", "context", "routes-query-modules.md"))
	if err != nil {
		t.Fatalf("read card: %v", err)
	}
	wantCard := `---
scope:
  - src/routes/**
tags:
  - database
---

# Routes Query Modules

Write the repo context here.
`
	if string(data) != wantCard {
		t.Fatalf("card content = %q, want %q", string(data), wantCard)
	}
}

func TestAgentsSyncPreviewFromContextCards(t *testing.T) {
	root := t.TempDir()
	garden := New(Options{Root: root})
	if err := garden.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	writeCard(t, root, "routes-query-modules", `---
scope:
  - src/routes/**
tags:
  - database
---

# Routes Query Modules

Use query modules.
`)

	change, err := garden.AgentsSync(AgentsSyncInput{Apply: false})
	if err != nil {
		t.Fatalf("AgentsSync preview returned error: %v", err)
	}
	if change.Applied {
		t.Fatal("preview should not be marked applied")
	}
	if change.Before != "" {
		t.Fatalf("preview before = %q, want empty missing AGENTS.md", change.Before)
	}
	assertRawAgentsContent(t, change.After)
	wantAfter := strings.Join([]string{
		agents.AgentsStartMarker,
		"### Garden Context",
		"",
		"Detailed agent context lives in `.garden/context/*.md`.",
		"",
		"Before editing a listed area, inspect the matching context card.",
		"",
		"Index:",
		agents.IndexStartMarker,
		"[Garden Context Index]|root:.garden/context",
		"|IMPORTANT:Before editing a listed area, inspect the matching context card",
		"|src/routes/**:.garden/context/routes-query-modules.md",
		agents.IndexEndMarker,
		agents.AgentsEndMarker,
	}, "\n") + "\n"
	if change.After != wantAfter {
		t.Fatalf("preview after = %q, want %q", change.After, wantAfter)
	}
	if len(change.Findings) != 0 {
		t.Fatalf("preview findings = %#v, want none", change.Findings)
	}
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("preview wrote AGENTS.md, stat err = %v", err)
	}
}

func TestAgentsSyncApplyWritesAGENTS(t *testing.T) {
	root := t.TempDir()
	garden := New(Options{Root: root})
	if err := garden.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	writeCard(t, root, "routes-query-modules", `---
scope:
  - src/routes/**
tags:
  - database
---

# Routes Query Modules

Use query modules.
`)

	change, err := garden.AgentsSync(AgentsSyncInput{Apply: true})
	if err != nil {
		t.Fatalf("AgentsSync apply returned error: %v", err)
	}
	if !change.Applied {
		t.Fatal("apply should be marked applied")
	}
	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(data) != change.After {
		t.Fatalf("written AGENTS.md did not match change.After:\nwritten:\n%s\nafter:\n%s", string(data), change.After)
	}
}

func TestLintReportsStaleAgentsIndex(t *testing.T) {
	root := t.TempDir()
	garden := New(Options{Root: root})
	if err := garden.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	writeCard(t, root, "routes-query-modules", `---
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
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(staleAgents), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	findings, err := garden.Lint()
	if err != nil {
		t.Fatalf("Lint returned error: %v", err)
	}
	assertAppFindings(t, findings, []agents.Finding{{
		Severity: "error",
		Code:     "stale-garden-index",
		Message:  "AGENTS.md Garden index is stale; run garden agents sync --apply",
	}})
}

func TestLintReportsInvalidContextCard(t *testing.T) {
	root := t.TempDir()
	garden := New(Options{Root: root})
	if err := garden.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	writeCard(t, root, "broken-card", `---
scope: src/routes/**
---

Use query modules.
`)
	block, err := agents.RenderBlock(nil)
	if err != nil {
		t.Fatalf("RenderBlock returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(block), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	findings, err := garden.Lint()
	if err != nil {
		t.Fatalf("Lint returned error: %v", err)
	}
	assertAppFindings(t, findings, []agents.Finding{{
		Severity: "error",
		Code:     "invalid-context-card",
		Message:  ".garden/context/broken-card.md: scope must be a list",
	}})
}

func TestRemoveDeletesContextCard(t *testing.T) {
	root := t.TempDir()
	garden := New(Options{Root: root})
	if err := garden.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if _, err := garden.NewCard(NewCardInput{Slug: "routes-query-modules", Scope: []string{"src/routes/**"}}); err != nil {
		t.Fatalf("NewCard returned error: %v", err)
	}

	if _, err := garden.RemoveCard("routes-query-modules"); err != nil {
		t.Fatalf("RemoveCard returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".garden", "context", "routes-query-modules.md")); !os.IsNotExist(err) {
		t.Fatalf("expected card removed, stat err = %v", err)
	}
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

func assertAppFindings(t *testing.T, got []agents.Finding, want []agents.Finding) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("findings = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("findings[%d] = %#v, want %#v; all findings = %#v", i, got[i], want[i], got)
		}
	}
}

func assertRawAgentsContent(t *testing.T, content string) {
	t.Helper()
	for _, marker := range []string{"--- AGENTS.md", "+++ AGENTS.md", "\n@@\n"} {
		if strings.Contains(content, marker) {
			t.Fatalf("change.After should be raw AGENTS.md content, found diff marker %q in:\n%s", marker, content)
		}
	}
}
