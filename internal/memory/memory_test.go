package memory

import (
	"strings"
	"testing"
	"time"
)

func TestNewRejectsInvalidMemoryOptions(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	gen := func() (string, error) { return "mem_1111111111", nil }

	tests := []struct {
		name    string
		text    string
		opts    Options
		wantErr string
	}{
		{
			name:    "empty memory",
			text:    "  ",
			opts:    Options{Scope: []string{"src/**"}},
			wantErr: "memory cannot be empty",
		},
		{
			name:    "missing scope or always",
			text:    "Use query modules.",
			opts:    Options{},
			wantErr: "use either --scope or --always",
		},
		{
			name:    "scope and always",
			text:    "Use query modules.",
			opts:    Options{Scope: []string{"src/**"}, Always: true},
			wantErr: "--scope and --always cannot be used together",
		},
		{
			name:    "invalid priority",
			text:    "Use query modules.",
			opts:    Options{Scope: []string{"src/**"}, Priority: "urgent"},
			wantErr: "invalid priority \"urgent\"; expected low, normal, or high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.text, tt.opts, now, nil, gen)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestNewDefaultsPriorityAndRegeneratesCollidingIDs(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 34, 56, 0, time.UTC)
	ids := []string{"mem_aaaaaaaaaa", "mem_bbbbbbbbbb"}
	gen := func() (string, error) {
		id := ids[0]
		ids = ids[1:]
		return id, nil
	}

	mem, err := New(
		" Route files should use query modules. ",
		Options{Scope: []string{"src/routes/**"}, Tags: []string{"database"}},
		now,
		map[string]bool{"mem_aaaaaaaaaa": true},
		gen,
	)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if mem.ID != "mem_bbbbbbbbbb" {
		t.Fatalf("ID = %q, want regenerated ID", mem.ID)
	}
	if mem.Memory != "Route files should use query modules." {
		t.Fatalf("Memory = %q", mem.Memory)
	}
	if mem.Priority != PriorityNormal {
		t.Fatalf("Priority = %q, want %q", mem.Priority, PriorityNormal)
	}
	if mem.CreatedAt != "2026-05-19T12:34:56.000Z" || mem.UpdatedAt != mem.CreatedAt {
		t.Fatalf("timestamps = created %q updated %q", mem.CreatedAt, mem.UpdatedAt)
	}
	if len(mem.Scope) != 1 || mem.Scope[0] != "src/routes/**" {
		t.Fatalf("Scope = %#v", mem.Scope)
	}
	if len(mem.Tags) != 1 || mem.Tags[0] != "database" {
		t.Fatalf("Tags = %#v", mem.Tags)
	}
}

func TestRandomIDShape(t *testing.T) {
	id, err := RandomID()
	if err != nil {
		t.Fatalf("RandomID returned error: %v", err)
	}
	if !ValidID(id) {
		t.Fatalf("RandomID() = %q, want mem_<10 hex chars>", id)
	}
}
