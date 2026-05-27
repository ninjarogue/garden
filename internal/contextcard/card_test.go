package contextcard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseReadsMarkdownCardFrontmatterAndBody(t *testing.T) {
	card, err := Parse(".garden/context/routes-query-modules.md", []byte(`---
scope:
  - src/routes/**
tags:
  - database
  - tenant-scoping
---

# Routes Query Modules

Route files should use query modules.
`))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if card.Slug != "routes-query-modules" {
		t.Fatalf("Slug = %q, want routes-query-modules", card.Slug)
	}
	if card.Path != ".garden/context/routes-query-modules.md" {
		t.Fatalf("Path = %q", card.Path)
	}
	assertStrings(t, card.Scope, []string{"src/routes/**"})
	assertStrings(t, card.Tags, []string{"database", "tenant-scoping"})
	if card.Body != "# Routes Query Modules\n\nRoute files should use query modules." {
		t.Fatalf("Body = %q", card.Body)
	}
}

func TestParseRejectsInvalidMarkdownCards(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
		wantErr string
	}{
		{
			name:    "missing frontmatter",
			path:    ".garden/context/routes-query-modules.md",
			content: "# Routes Query Modules\n\nUse query modules.\n",
			wantErr: "YAML frontmatter is required",
		},
		{
			name:    "missing scope",
			path:    ".garden/context/routes-query-modules.md",
			content: "---\nscope: []\n---\n\nUse query modules.\n",
			wantErr: "scope must include at least one glob",
		},
		{
			name:    "scope must be list",
			path:    ".garden/context/routes-query-modules.md",
			content: "---\nscope: src/routes/**\n---\n\nUse query modules.\n",
			wantErr: "scope must be a list",
		},
		{
			name:    "placeholder scope",
			path:    ".garden/context/routes-query-modules.md",
			content: "---\nscope:\n  - CHANGE_ME\n---\n\nUse query modules.\n",
			wantErr: "scope cannot contain CHANGE_ME",
		},
		{
			name:    "invalid scope glob",
			path:    ".garden/context/routes-query-modules.md",
			content: "---\nscope:\n  - internal/[*.go\n---\n\nUse query modules.\n",
			wantErr: `invalid scope glob "internal/[*.go"`,
		},
		{
			name:    "tags must be list",
			path:    ".garden/context/routes-query-modules.md",
			content: "---\nscope:\n  - src/routes/**\ntags: database\n---\n\nUse query modules.\n",
			wantErr: "tags must be a list",
		},
		{
			name:    "empty body",
			path:    ".garden/context/routes-query-modules.md",
			content: "---\nscope:\n  - src/routes/**\n---\n\n  \n",
			wantErr: "body cannot be empty",
		},
		{
			name:    "invalid slug",
			path:    ".garden/context/Routes_Query_Modules.md",
			content: "---\nscope:\n  - src/routes/**\n---\n\nUse query modules.\n",
			wantErr: "invalid card slug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.path, []byte(tt.content))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParseAllowsCompactIndexDelimitersInHumanOnlyTags(t *testing.T) {
	card, err := Parse(".garden/context/routes-query-modules.md", []byte(`---
scope:
  - src/routes/**
tags:
  - database,tenant
  - "{tenant|scope}"
---

Use query modules.
`))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	assertStrings(t, card.Tags, []string{"database,tenant", "{tenant|scope}"})
}

func TestStoreCreatesCardWithYAMLSensitiveGlobScope(t *testing.T) {
	root := t.TempDir()
	store := NewStore(root)

	card, err := store.Create(CreateInput{
		Slug:  "global-background",
		Scope: []string{"**/*", "*.go"},
		Tags:  []string{"workflow"},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	assertStrings(t, card.Scope, []string{"**/*", "*.go"})

	data, err := os.ReadFile(filepath.Join(root, ".garden", "context", "global-background.md"))
	if err != nil {
		t.Fatalf("read card: %v", err)
	}
	wantCard := `---
scope:
  - '**/*'
  - '*.go'
tags:
  - workflow
---

# Global Background

Write the repo context here.
`
	if string(data) != wantCard {
		t.Fatalf("card content = %q, want %q", string(data), wantCard)
	}
	parsed, err := Parse(".garden/context/global-background.md", data)
	if err != nil {
		t.Fatalf("Parse created card returned error: %v\n%s", err, string(data))
	}
	assertStrings(t, parsed.Scope, []string{"**/*", "*.go"})
}

func TestStoreCreatePreservesHumanOnlyTagsWithCompactIndexDelimiters(t *testing.T) {
	root := t.TempDir()
	store := NewStore(root)

	card, err := store.Create(CreateInput{
		Slug:  "special-tags",
		Scope: []string{"src/**"},
		Tags:  []string{"database,tenant", "{tenant|scope}"},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	assertStrings(t, card.Tags, []string{"database,tenant", "{tenant|scope}"})

	data, err := os.ReadFile(filepath.Join(root, ".garden", "context", "special-tags.md"))
	if err != nil {
		t.Fatalf("read card: %v", err)
	}
	parsed, err := Parse(".garden/context/special-tags.md", data)
	if err != nil {
		t.Fatalf("Parse created card returned error: %v\n%s", err, string(data))
	}
	assertStrings(t, parsed.Tags, []string{"database,tenant", "{tenant|scope}"})
}

func TestStoreCreateRejectsInvalidScopeGlob(t *testing.T) {
	root := t.TempDir()
	store := NewStore(root)

	_, err := store.Create(CreateInput{
		Slug:  "broken-scope",
		Scope: []string{"internal/[*.go"},
	})
	if err == nil {
		t.Fatal("expected invalid scope glob error")
	}
	wantErr := `invalid scope glob "internal/[*.go"`
	if !strings.Contains(err.Error(), wantErr) {
		t.Fatalf("error = %q, want substring %q", err.Error(), wantErr)
	}
}

func TestStoreCreateRejectsDuplicateCardSlug(t *testing.T) {
	root := t.TempDir()
	store := NewStore(root)

	if _, err := store.Create(CreateInput{Slug: "routes-query-modules", Scope: []string{"src/routes/**"}}); err != nil {
		t.Fatalf("first Create returned error: %v", err)
	}

	_, err := store.Create(CreateInput{Slug: "routes-query-modules", Scope: []string{"src/routes/**"}})
	if err == nil {
		t.Fatal("expected duplicate card error")
	}
	wantErr := "context card already exists: .garden/context/routes-query-modules.md"
	if err.Error() != wantErr {
		t.Fatalf("error = %q, want %q", err.Error(), wantErr)
	}
}

func TestStoreCreatesContextDirectoryAndCardTemplate(t *testing.T) {
	root := t.TempDir()
	store := NewStore(root)

	if err := store.Init(); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".garden", "context")); err != nil {
		t.Fatalf("expected context directory: %v", err)
	}

	card, err := store.Create(CreateInput{
		Slug:  "routes-query-modules",
		Scope: []string{"src/routes/**"},
		Tags:  []string{"database"},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
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

func assertStrings(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("strings = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("strings[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
