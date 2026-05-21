package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aric/garden/internal/app"
	"github.com/aric/garden/internal/storage"
)

func TestRememberAndPackCommands(t *testing.T) {
	rootDir := t.TempDir()
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	ids := []string{"mem_1111111111"}
	garden := app.New(app.Options{
		Store: storage.NewJSONStore(rootDir),
		Now:   func() time.Time { return now },
		IDGenerator: func() (string, error) {
			id := ids[0]
			ids = ids[1:]
			return id, nil
		},
	})

	if _, _, err := execute(garden, "init"); err != nil {
		t.Fatalf("init returned error: %v", err)
	}
	out, _, err := execute(garden,
		"remember",
		"Route files should not import DB clients directly; query modules enforce tenant scoping.",
		"--scope", "src/routes/**",
		"--tag", "database",
		"--priority", "high",
	)
	if err != nil {
		t.Fatalf("remember returned error: %v", err)
	}
	if !strings.Contains(out, "mem_1111111111") {
		t.Fatalf("remember output = %q, want memory ID", out)
	}

	out, _, err = execute(garden,
		"pack",
		"--path", "src/routes/api/users.ts",
		"--task", "add user database endpoint",
		"--explain",
	)
	if err != nil {
		t.Fatalf("pack --explain returned error: %v", err)
	}
	for _, want := range []string{
		"## Why These Memories",
		"mem_1111111111 selected:",
		"- scope matched `src/routes/**` (+40)",
		"- tag `database` matched task (+8)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("pack --explain output missing %q:\n%s", want, out)
		}
	}

	out, _, err = execute(garden,
		"pack",
		"--path", "src/routes/api/users.ts",
		"--task", "add user database endpoint",
	)
	if err != nil {
		t.Fatalf("pack returned error: %v", err)
	}
	for _, want := range []string{
		"<garden_context_pack>",
		"Path: `src/routes/api/users.ts`",
		"Task: add user database endpoint",
		"- Route files should not import DB clients directly; query modules enforce tenant scoping.",
		"</garden_context_pack>",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("pack output missing %q:\n%s", want, out)
		}
	}
}

func TestAgentsCommandsWireFlagsAndPreviewByDefault(t *testing.T) {
	rootDir := t.TempDir()
	garden := app.New(app.Options{Root: rootDir, Store: storage.NewJSONStore(rootDir)})
	if _, _, err := execute(garden, "init"); err != nil {
		t.Fatalf("init returned error: %v", err)
	}

	out, _, err := execute(garden,
		"agents", "update",
		"--purpose", "Local context and memory router for coding agents.",
		"--test", "mise exec -- go test ./...",
		"--map", "internal/cmd: Cobra command parsing and output wiring",
	)
	if err != nil {
		t.Fatalf("agents update preview returned error: %v", err)
	}
	for _, want := range []string{"--- AGENTS.md", "+<!-- garden:agents:start -->", "Preview only. Re-run with --apply to write AGENTS.md."} {
		if !strings.Contains(out, want) {
			t.Fatalf("agents update preview missing %q:\n%s", want, out)
		}
	}
	if _, err := os.Stat(filepath.Join(rootDir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("preview wrote AGENTS.md, stat err = %v", err)
	}

	out, _, err = execute(garden,
		"agents", "update",
		"--purpose", "Local context and memory router for coding agents.",
		"--test", "mise exec -- go test ./...",
		"--map", "internal/cmd: Cobra command parsing and output wiring",
		"--apply",
	)
	if err != nil {
		t.Fatalf("agents update apply returned error: %v", err)
	}
	if !strings.Contains(out, "Applied AGENTS.md update.") {
		t.Fatalf("agents update apply output = %q", out)
	}

	out, _, err = execute(garden, "agents", "sync")
	if err != nil {
		t.Fatalf("agents sync preview returned error: %v", err)
	}
	if !strings.Contains(out, "Preview only. Re-run with --apply to write AGENTS.md.") {
		t.Fatalf("agents sync preview output = %q", out)
	}

	out, _, err = execute(garden, "agents", "lint")
	if err != nil {
		t.Fatalf("agents lint returned error: %v", err)
	}
	if !strings.Contains(out, "AGENTS.md lint passed.") {
		t.Fatalf("agents lint output = %q", out)
	}
}

func TestCommandValidationReturnsActionableErrors(t *testing.T) {
	garden := app.New(app.Options{Store: storage.NewJSONStore(t.TempDir())})

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "remember requires scope or always", args: []string{"remember", "Use query modules."}, wantErr: "use either --scope or --always"},
		{name: "pack requires path", args: []string{"pack", "--task", "add endpoint"}, wantErr: "--path is required"},
		{name: "pack requires task", args: []string{"pack", "--path", "src/file.go"}, wantErr: "--task is required"},
		{name: "agents update parses map before app orchestration", args: []string{"agents", "update", "--map", "internal/cmd without separator"}, wantErr: "--map: expected <path>: <description>"},
		{name: "agents update parses doc before app orchestration", args: []string{"agents", "update", "--doc", "README.md without separator"}, wantErr: "--doc: expected <path>: <description>"},
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
