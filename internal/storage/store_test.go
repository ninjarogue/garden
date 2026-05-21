package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aric/garden/internal/memory"
)

func TestInitCreatesVersionedMemoryFile(t *testing.T) {
	root := t.TempDir()
	store := NewJSONStore(root)

	if err := store.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	path := filepath.Join(root, ".garden", "memories.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected memories file: %v", err)
	}

	doc, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if doc.Version != 1 {
		t.Fatalf("Version = %d, want 1", doc.Version)
	}
	if len(doc.Memories) != 0 {
		t.Fatalf("Memories = %#v, want empty", doc.Memories)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	root := t.TempDir()
	store := NewJSONStore(root)
	if err := store.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	doc := memory.Document{
		Version: 1,
		Memories: []memory.Memory{{
			ID:        "mem_1111111111",
			Memory:    "Use query modules.",
			Scope:     []string{"src/routes/**"},
			Priority:  memory.PriorityNormal,
			CreatedAt: "2026-05-19T12:00:00.000Z",
			UpdatedAt: "2026-05-19T12:00:00.000Z",
		}},
	}
	if err := store.Save(doc); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(loaded.Memories) != 1 || loaded.Memories[0].ID != "mem_1111111111" {
		t.Fatalf("loaded = %#v", loaded)
	}
}
