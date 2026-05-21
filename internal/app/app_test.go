package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aric/garden/internal/agents"
	"github.com/aric/garden/internal/memory"
	"github.com/aric/garden/internal/storage"
)

func TestAgentsUpdatePreviewDoesNotWriteAndApplyWritesAGENTS(t *testing.T) {
	root := t.TempDir()
	store := storage.NewJSONStore(root)
	if err := store.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if err := store.Save(memory.Document{Version: memory.Version, Memories: []memory.Memory{{
		ID:       "mem_1111111111",
		Memory:   "Route files should not import DB clients directly.",
		Scope:    []string{"src/routes/**"},
		Tags:     []string{"database"},
		Priority: memory.PriorityNormal,
	}}}); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	garden := New(Options{Root: root, Store: store})

	input := AgentsUpdateInput{
		Purpose: "Local context and memory router for coding agents.",
		Test:    []string{"mise exec -- go test ./..."},
		Map:     []agents.Entry{{Path: "internal/app", Text: "use case orchestration"}},
		Apply:   false,
	}
	change, err := garden.AgentsUpdate(input)
	if err != nil {
		t.Fatalf("AgentsUpdate preview returned error: %v", err)
	}
	if change.Applied {
		t.Fatal("preview should not be marked applied")
	}
	if change.Before != "" {
		t.Fatalf("preview before = %q, want empty missing AGENTS.md", change.Before)
	}
	if !strings.Contains(change.After, "<!-- garden:agents:start -->") || !strings.Contains(change.After, "|src/routes/**:{database,mem_1111111111}") {
		t.Fatalf("preview after missing rendered block/index:\n%s", change.After)
	}
	assertRawAgentsContent(t, change.After)
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("preview wrote AGENTS.md, stat err = %v", err)
	}

	input.Apply = true
	change, err = garden.AgentsUpdate(input)
	if err != nil {
		t.Fatalf("AgentsUpdate apply returned error: %v", err)
	}
	if !change.Applied {
		t.Fatal("apply should be marked applied")
	}
	assertRawAgentsContent(t, change.After)
	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	content := string(data)
	if content != change.After {
		t.Fatalf("written AGENTS.md did not match change.After:\nwritten:\n%s\nafter:\n%s", content, change.After)
	}
	for _, want := range []string{"## Garden Agent Context", "Local context and memory router", "|src/routes/**:{database,mem_1111111111}"} {
		if !strings.Contains(content, want) {
			t.Fatalf("AGENTS.md missing %q:\n%s", want, content)
		}
	}
}

func TestAgentsSyncPreviewAndApplyRefreshOnlyIndex(t *testing.T) {
	root := t.TempDir()
	store := storage.NewJSONStore(root)
	if err := store.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if err := store.Save(memory.Document{Version: memory.Version, Memories: []memory.Memory{{
		ID:       "mem_2222222222",
		Memory:   "Retrieval ranks by budget and lexical matches.",
		Scope:    []string{"internal/retrieval/**"},
		Tags:     []string{"ranking"},
		Priority: memory.PriorityHigh,
	}}}); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	agentsPath := filepath.Join(root, "AGENTS.md")
	original := strings.Join([]string{
		"# Human Rules",
		agents.AgentsStartMarker,
		"## Garden Agent Context",
		"### Project Purpose",
		"Keep Garden small.",
		agents.IndexStartMarker,
		"[Garden Memory Index]|root:.garden/memories.json",
		"|old/**:{old,mem_0000000000}",
		agents.IndexEndMarker,
		agents.AgentsEndMarker,
		"Do not remove this.",
	}, "\n") + "\n"
	if err := os.WriteFile(agentsPath, []byte(original), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	garden := New(Options{Root: root, Store: store})

	change, err := garden.AgentsSync(AgentsSyncInput{Apply: false})
	if err != nil {
		t.Fatalf("AgentsSync preview returned error: %v", err)
	}
	if change.Applied {
		t.Fatal("preview should not be marked applied")
	}
	if change.Before != original {
		t.Fatalf("sync before changed:\n%s", change.Before)
	}
	if !strings.Contains(change.After, "|internal/retrieval/**:{ranking,mem_2222222222}") {
		t.Fatalf("sync after missing refreshed index:\n%s", change.After)
	}
	assertRawAgentsContent(t, change.After)
	data, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(data) != original {
		t.Fatalf("sync preview wrote AGENTS.md:\n%s", string(data))
	}

	change, err = garden.AgentsSync(AgentsSyncInput{Apply: true})
	if err != nil {
		t.Fatalf("AgentsSync apply returned error: %v", err)
	}
	if !change.Applied {
		t.Fatal("apply should be marked applied")
	}
	assertRawAgentsContent(t, change.After)
	data, err = os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	content := string(data)
	if content != change.After {
		t.Fatalf("written AGENTS.md did not match change.After:\nwritten:\n%s\nafter:\n%s", content, change.After)
	}
	for _, want := range []string{"# Human Rules", "Keep Garden small.", "Do not remove this.", "|internal/retrieval/**:{ranking,mem_2222222222}"} {
		if !strings.Contains(content, want) {
			t.Fatalf("synced AGENTS.md missing %q:\n%s", want, content)
		}
	}
	if strings.Contains(content, "old/**") {
		t.Fatalf("synced AGENTS.md kept old index:\n%s", content)
	}
}

func TestAgentsSyncRequiresExistingGardenBlock(t *testing.T) {
	root := t.TempDir()
	store := storage.NewJSONStore(root)
	if err := store.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("# Human Rules\n"), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	garden := New(Options{Root: root, Store: store})

	_, err := garden.AgentsSync(AgentsSyncInput{Apply: false})
	if err == nil {
		t.Fatal("expected missing Garden agents block error")
	}
	if !strings.Contains(err.Error(), "Garden agents block is missing") {
		t.Fatalf("error = %q, want missing block message", err.Error())
	}
}

func TestAgentsUpdateRejectsReservedGardenMarkers(t *testing.T) {
	root := t.TempDir()
	store := storage.NewJSONStore(root)
	if err := store.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	garden := New(Options{Root: root, Store: store})

	_, err := garden.AgentsUpdate(AgentsUpdateInput{Purpose: "break " + agents.AgentsEndMarker})
	if err == nil {
		t.Fatal("expected reserved marker error")
	}
	if !strings.Contains(err.Error(), "reserved Garden marker") {
		t.Fatalf("error = %q, want reserved marker message", err.Error())
	}
	if _, statErr := os.Stat(filepath.Join(root, "AGENTS.md")); !os.IsNotExist(statErr) {
		t.Fatalf("invalid update wrote AGENTS.md, stat err = %v", statErr)
	}
}

func TestAgentsLintUsesStoredMemoriesForFullBodyFindings(t *testing.T) {
	root := t.TempDir()
	store := storage.NewJSONStore(root)
	if err := store.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	memoryBody := "Route files should not import DB clients directly."
	if err := store.Save(memory.Document{Version: memory.Version, Memories: []memory.Memory{{
		ID:       "mem_3333333333",
		Memory:   memoryBody,
		Scope:    []string{"src/routes/**"},
		Priority: memory.PriorityNormal,
	}}}); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	doc := agents.AgentsStartMarker + "\n## Garden Agent Context\n" + memoryBody + "\n" + agents.AgentsEndMarker + "\n"
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(doc), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	garden := New(Options{Root: root, Store: store})

	findings, err := garden.AgentsLint()
	if err != nil {
		t.Fatalf("AgentsLint returned error: %v", err)
	}
	if !appHasFinding(findings, "full-memory-body") {
		t.Fatalf("findings missing full-memory-body: %#v", findings)
	}
}

func appHasFinding(findings []agents.Finding, code string) bool {
	for _, finding := range findings {
		if finding.Code == code {
			return true
		}
	}
	return false
}

func assertRawAgentsContent(t *testing.T, content string) {
	t.Helper()
	for _, marker := range []string{"--- AGENTS.md", "+++ AGENTS.md", "\n@@\n"} {
		if strings.Contains(content, marker) {
			t.Fatalf("change.After should be raw AGENTS.md content, found diff marker %q in:\n%s", marker, content)
		}
	}
}
