package agents

import (
	"strings"
	"testing"
)

func TestUpsertBlockReplacesOrAppendsOnlyGardenAgentsBlock(t *testing.T) {
	newBlock := AgentsStartMarker + "\nnew garden block\n" + AgentsEndMarker + "\n"
	tests := []struct {
		name string
		doc  string
		want string
	}{
		{
			name: "replaces existing block and preserves surrounding content",
			doc: strings.Join([]string{
				"# Human Rules",
				"Keep this.",
				"<!-- other-tool:start -->",
				"other tool block",
				"<!-- other-tool:end -->",
				AgentsStartMarker,
				"old garden block",
				AgentsEndMarker,
				"After text.",
			}, "\n") + "\n",
			want: strings.Join([]string{
				"# Human Rules",
				"Keep this.",
				"<!-- other-tool:start -->",
				"other tool block",
				"<!-- other-tool:end -->",
				AgentsStartMarker,
				"new garden block",
				AgentsEndMarker,
				"",
				"After text.",
			}, "\n") + "\n",
		},
		{
			name: "appends block when absent",
			doc:  "# Human Rules\nKeep this.\n",
			want: strings.Join([]string{
				"# Human Rules",
				"Keep this.",
				"",
				AgentsStartMarker,
				"new garden block",
				AgentsEndMarker,
			}, "\n") + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpsertBlock(tt.doc, newBlock)
			if err != nil {
				t.Fatalf("UpsertBlock returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("upserted doc = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUpsertBlockRejectsMalformedGardenMarkers(t *testing.T) {
	newBlock := AgentsStartMarker + "\nnew\n" + AgentsEndMarker + "\n"
	tests := []struct {
		name string
		doc  string
	}{
		{name: "start without end", doc: "before\n" + AgentsStartMarker + "\nmissing end\n"},
		{name: "end without start", doc: "before\n" + AgentsEndMarker + "\n"},
		{name: "duplicate start", doc: AgentsStartMarker + "\n" + AgentsStartMarker + "\n" + AgentsEndMarker + "\n"},
		{name: "duplicate end", doc: AgentsStartMarker + "\n" + AgentsEndMarker + "\n" + AgentsEndMarker + "\n"},
		{name: "reversed", doc: AgentsEndMarker + "\n" + AgentsStartMarker + "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UpsertBlock(tt.doc, newBlock)
			if err == nil {
				t.Fatal("expected malformed marker error")
			}
			if !strings.Contains(err.Error(), "malformed Garden agents markers") {
				t.Fatalf("error = %q, want malformed marker message", err.Error())
			}
		})
	}
}

func TestSyncIndexCreatesManagedBlockWhenAbsent(t *testing.T) {
	got, err := SyncIndex("# Human Rules\n", []IndexCard{{
		Path:  ".garden/context/routes-query-modules.md",
		Kind:  "rule",
		Scope: []string{"src/routes/**"},
		Tags:  []string{"database"},
	}})
	if err != nil {
		t.Fatalf("SyncIndex returned error: %v", err)
	}

	want := strings.Join([]string{
		"# Human Rules",
		"",
		AgentsStartMarker,
		"### Garden Context",
		"",
		"Detailed agent context lives in `.garden/context/*.md`.",
		"",
		"Before editing a listed area, inspect the matching context card.",
		"",
		"Index:",
		IndexStartMarker,
		"[Garden Context Index]|root:.garden/context",
		"|IMPORTANT:Before editing a listed area, inspect the matching context card",
		"|src/routes/**:{rule,database,.garden/context/routes-query-modules.md}",
		IndexEndMarker,
		AgentsEndMarker,
	}, "\n") + "\n"
	if got != want {
		t.Fatalf("synced doc = %q, want %q", got, want)
	}
}

func TestSyncIndexAddsMissingIndexInsideExistingGardenBlock(t *testing.T) {
	doc := strings.Join([]string{
		"# Human Rules",
		AgentsStartMarker,
		"### Garden Context",
		"Human-edited managed prose.",
		AgentsEndMarker,
	}, "\n") + "\n"

	got, err := SyncIndex(doc, []IndexCard{{
		Path:  ".garden/context/context-card-format.md",
		Kind:  "workflow",
		Scope: []string{"internal/contextcard/**"},
		Tags:  []string{"frontmatter"},
	}})
	if err != nil {
		t.Fatalf("SyncIndex returned error: %v", err)
	}

	want := strings.Join([]string{
		"# Human Rules",
		AgentsStartMarker,
		"### Garden Context",
		"Human-edited managed prose.",
		"",
		IndexStartMarker,
		"[Garden Context Index]|root:.garden/context",
		"|IMPORTANT:Before editing a listed area, inspect the matching context card",
		"|internal/contextcard/**:{workflow,frontmatter,.garden/context/context-card-format.md}",
		IndexEndMarker,
		AgentsEndMarker,
	}, "\n") + "\n"
	if got != want {
		t.Fatalf("synced doc = %q, want %q", got, want)
	}
}

func TestSyncIndexDoesNotInsertBlankLineBeforeAgentsEndMarker(t *testing.T) {
	doc := strings.Join([]string{
		AgentsStartMarker,
		"### Garden Context",
		IndexStartMarker,
		"[Garden Context Index]|root:.garden/context",
		"|old/**:{old,.garden/context/old.md}",
		IndexEndMarker,
		AgentsEndMarker,
	}, "\n") + "\n"

	got, err := SyncIndex(doc, []IndexCard{{
		Path:  ".garden/context/routes-query-modules.md",
		Kind:  "rule",
		Scope: []string{"src/routes/**"},
	}})
	if err != nil {
		t.Fatalf("SyncIndex returned error: %v", err)
	}

	want := strings.Join([]string{
		AgentsStartMarker,
		"### Garden Context",
		IndexStartMarker,
		"[Garden Context Index]|root:.garden/context",
		"|IMPORTANT:Before editing a listed area, inspect the matching context card",
		"|src/routes/**:{rule,.garden/context/routes-query-modules.md}",
		IndexEndMarker,
		AgentsEndMarker,
	}, "\n") + "\n"
	if got != want {
		t.Fatalf("synced doc = %q, want %q", got, want)
	}
}
