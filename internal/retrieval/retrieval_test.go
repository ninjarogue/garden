package retrieval

import (
	"testing"

	"github.com/aric/garden/internal/memory"
)

func TestSelectFiltersCandidatesAndRanksByDeterministicScore(t *testing.T) {
	memories := []memory.Memory{
		{
			ID:        "mem_3333333333",
			Memory:    "Database setup applies everywhere.",
			Always:    true,
			Tags:      []string{"database"},
			Priority:  memory.PriorityHigh,
			UpdatedAt: "2026-05-19T10:00:00.000Z",
		},
		{
			ID:        "mem_1111111111",
			Memory:    "Use query modules for database user endpoints.",
			Scope:     []string{"src/routes/**"},
			Tags:      []string{"database"},
			Priority:  memory.PriorityNormal,
			UpdatedAt: "2026-05-19T09:00:00.000Z",
		},
		{
			ID:        "mem_2222222222",
			Memory:    "Docs pages use markdown frontmatter.",
			Scope:     []string{"docs/**"},
			Tags:      []string{"database"},
			Priority:  memory.PriorityHigh,
			UpdatedAt: "2026-05-19T11:00:00.000Z",
		},
	}

	results, err := Select(memories, Query{
		Path:     "src/routes/api/users.ts",
		Task:     "add user endpoint database",
		Max:      8,
		MaxChars: 4000,
	})
	if err != nil {
		t.Fatalf("Select returned error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Memory.ID != "mem_1111111111" {
		t.Fatalf("first result = %q, want scoped route memory", results[0].Memory.ID)
	}
	if results[1].Memory.ID != "mem_3333333333" {
		t.Fatalf("second result = %q, want always memory", results[1].Memory.ID)
	}
}

func TestSelectUsesUpdatedAtAndIDTieBreakers(t *testing.T) {
	memories := []memory.Memory{
		{
			ID:        "mem_bbbbbbbbbb",
			Memory:    "Use route modules.",
			Scope:     []string{"src/**"},
			Priority:  memory.PriorityNormal,
			UpdatedAt: "2026-05-19T10:00:00.000Z",
		},
		{
			ID:        "mem_aaaaaaaaaa",
			Memory:    "Use route modules.",
			Scope:     []string{"src/**"},
			Priority:  memory.PriorityNormal,
			UpdatedAt: "2026-05-19T10:00:00.000Z",
		},
		{
			ID:        "mem_cccccccccc",
			Memory:    "Use route modules.",
			Scope:     []string{"src/**"},
			Priority:  memory.PriorityNormal,
			UpdatedAt: "2026-05-19T11:00:00.000Z",
		},
	}

	results, err := Select(memories, Query{Path: "src/file.go", Task: "route", Max: 8, MaxChars: 4000})
	if err != nil {
		t.Fatalf("Select returned error: %v", err)
	}

	want := []string{"mem_cccccccccc", "mem_aaaaaaaaaa", "mem_bbbbbbbbbb"}
	for i, id := range want {
		if results[i].Memory.ID != id {
			t.Fatalf("results[%d] = %q, want %q", i, results[i].Memory.ID, id)
		}
	}
}

func TestSelectHonorsMaxAndMaxCharsBudget(t *testing.T) {
	memories := []memory.Memory{
		{ID: "mem_1111111111", Memory: "12345", Scope: []string{"src/**"}, Priority: memory.PriorityNormal},
		{ID: "mem_2222222222", Memory: "123456", Scope: []string{"src/**"}, Priority: memory.PriorityNormal},
		{ID: "mem_3333333333", Memory: "123", Scope: []string{"src/**"}, Priority: memory.PriorityNormal},
	}

	results, err := Select(memories, Query{Path: "src/file.go", Task: "anything", Max: 2, MaxChars: 10})
	if err != nil {
		t.Fatalf("Select returned error: %v", err)
	}
	if len(results) != 1 || results[0].Memory.ID != "mem_1111111111" {
		t.Fatalf("results = %#v, want only first memory before char budget stops selection", results)
	}
}

func TestSelectIncludesSelectionReasons(t *testing.T) {
	memories := []memory.Memory{{
		ID:        "mem_1111111111",
		Memory:    "Use query modules for database endpoints.",
		Scope:     []string{"src/routes/**"},
		Tags:      []string{"database"},
		Priority:  memory.PriorityHigh,
		UpdatedAt: "2026-05-19T09:00:00.000Z",
	}}

	results, err := Select(memories, Query{Path: "src/routes/api/users.ts", Task: "database", Max: 8, MaxChars: 4000})
	if err != nil {
		t.Fatalf("Select returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}

	want := []Reason{
		{Text: "scope matched `src/routes/**`", Points: 40},
		{Text: "memory matched task token `database`", Points: 3},
		{Text: "tag `database` matched task", Points: 8},
		{Text: "priority high", Points: 10},
	}
	if len(results[0].Reasons) != len(want) {
		t.Fatalf("Reasons = %#v, want %#v", results[0].Reasons, want)
	}
	for i := range want {
		if results[0].Reasons[i] != want[i] {
			t.Fatalf("Reasons[%d] = %#v, want %#v", i, results[0].Reasons[i], want[i])
		}
	}
}

func TestTokenizeSplitsOnNonAlphanumericBoundaries(t *testing.T) {
	tokens := Tokenize("Add user_endpoint: database/routes!")
	want := []string{"add", "user", "endpoint", "database", "routes"}
	if len(tokens) != len(want) {
		t.Fatalf("tokens = %#v", tokens)
	}
	for i := range want {
		if tokens[i] != want[i] {
			t.Fatalf("tokens[%d] = %q, want %q", i, tokens[i], want[i])
		}
	}
}
