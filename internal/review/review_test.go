package review

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuildReportMatchesChangedFilesToCards(t *testing.T) {
	report, err := BuildReport(Input{
		ChangedPaths: []string{"./internal/cmd/root.go"},
		Cards: []Card{{
			Path:  ".garden/context/app-layer-architecture.md",
			Scope: []string{"internal/app/**", "internal/cmd/**"},
			Body: `# App Layer Architecture

Keep commands thin.

## Verification

Run:

` + "```sh" + `
env GOCACHE=/tmp/garden-go-build go test ./...
` + "```" + `

## Notes

Other guidance.`,
		}},
	})
	if err != nil {
		t.Fatalf("BuildReport returned error: %v", err)
	}

	want := Report{ChangedFiles: []ChangedFile{{
		Path: "internal/cmd/root.go",
		Cards: []MatchedCard{{
			Path:         ".garden/context/app-layer-architecture.md",
			MatchedScope: "internal/cmd/**",
			Verification: "Run:\n\n```sh\nenv GOCACHE=/tmp/garden-go-build go test ./...\n```",
		}},
	}}}
	assertReport(t, report, want)
}

func TestBuildReportNormalizesChangedPathSeparators(t *testing.T) {
	report, err := BuildReport(Input{
		ChangedPaths: []string{`internal\cmd\root.go`},
		Cards: []Card{{
			Path:  ".garden/context/app-layer-architecture.md",
			Scope: []string{"internal/cmd/**"},
			Body:  "App guidance.",
		}},
	})
	if err != nil {
		t.Fatalf("BuildReport returned error: %v", err)
	}

	want := Report{ChangedFiles: []ChangedFile{{
		Path:  "internal/cmd/root.go",
		Cards: []MatchedCard{{Path: ".garden/context/app-layer-architecture.md", MatchedScope: "internal/cmd/**"}},
	}}}
	assertReport(t, report, want)
}

func TestBuildReportIsDeterministicAcrossInputOrder(t *testing.T) {
	report, err := BuildReport(Input{
		ChangedPaths: []string{"src/z.go", "src/a_test.go", "src/a.go"},
		Cards: []Card{
			{Path: ".garden/context/tests.md", Scope: []string{"**/*_test.go"}, Body: "Test guidance."},
			{Path: ".garden/context/source.md", Scope: []string{"src/**"}, Body: "Source guidance."},
		},
	})
	if err != nil {
		t.Fatalf("BuildReport returned error: %v", err)
	}

	want := Report{
		ChangedFiles: []ChangedFile{
			{Path: "src/a.go", Cards: []MatchedCard{{Path: ".garden/context/source.md", MatchedScope: "src/**"}}},
			{Path: "src/a_test.go", Cards: []MatchedCard{
				{Path: ".garden/context/source.md", MatchedScope: "src/**"},
				{Path: ".garden/context/tests.md", MatchedScope: "**/*_test.go"},
			}},
			{Path: "src/z.go", Cards: []MatchedCard{{Path: ".garden/context/source.md", MatchedScope: "src/**"}}},
		},
		Warnings: []Warning{{
			Path:    "src/a_test.go",
			Code:    "verification-surface-changed",
			Message: "changed test file",
		}},
	}
	assertReport(t, report, want)
}

func TestBuildReportRepresentsUnmatchedChangedFiles(t *testing.T) {
	report, err := BuildReport(Input{
		ChangedPaths: []string{"README.md"},
		Cards:        []Card{{Path: ".garden/context/app.md", Scope: []string{"internal/**"}, Body: "App guidance."}},
	})
	if err != nil {
		t.Fatalf("BuildReport returned error: %v", err)
	}

	want := Report{ChangedFiles: []ChangedFile{{Path: "README.md"}}}
	assertReport(t, report, want)
}

func TestBuildReportDoubleStarMatchesNestedPaths(t *testing.T) {
	report, err := BuildReport(Input{
		ChangedPaths: []string{"internal/cmd/sub/root.go", "src/deep/a_test.go"},
		Cards: []Card{
			{Path: ".garden/context/app.md", Scope: []string{"internal/cmd/**"}, Body: "App guidance."},
			{Path: ".garden/context/tests.md", Scope: []string{"**/*_test.go"}, Body: "Test guidance."},
		},
	})
	if err != nil {
		t.Fatalf("BuildReport returned error: %v", err)
	}

	want := Report{
		ChangedFiles: []ChangedFile{
			{Path: "internal/cmd/sub/root.go", Cards: []MatchedCard{{Path: ".garden/context/app.md", MatchedScope: "internal/cmd/**"}}},
			{Path: "src/deep/a_test.go", Cards: []MatchedCard{{Path: ".garden/context/tests.md", MatchedScope: "**/*_test.go"}}},
		},
		Warnings: []Warning{{Path: "src/deep/a_test.go", Code: "verification-surface-changed", Message: "changed test file"}},
	}
	assertReport(t, report, want)
}

func TestBuildReportWarnsForVerificationSurfaces(t *testing.T) {
	report, err := BuildReport(Input{ChangedPaths: []string{
		".github/workflows/test.yml",
		".garden/context/app-layer-architecture.md",
		".golangci.yml",
		"Makefile",
		"go.mod",
		"internal/cmd/root_test.go",
	}})
	if err != nil {
		t.Fatalf("BuildReport returned error: %v", err)
	}

	want := Report{
		ChangedFiles: []ChangedFile{
			{Path: ".garden/context/app-layer-architecture.md"},
			{Path: ".github/workflows/test.yml"},
			{Path: ".golangci.yml"},
			{Path: "Makefile"},
			{Path: "go.mod"},
			{Path: "internal/cmd/root_test.go"},
		},
		Warnings: []Warning{
			{Path: ".garden/context/app-layer-architecture.md", Code: "verification-surface-changed", Message: "changed Garden context card"},
			{Path: ".github/workflows/test.yml", Code: "verification-surface-changed", Message: "changed GitHub workflow"},
			{Path: ".golangci.yml", Code: "verification-surface-changed", Message: "changed lint or format config"},
			{Path: "Makefile", Code: "verification-surface-changed", Message: "changed build config"},
			{Path: "go.mod", Code: "verification-surface-changed", Message: "changed build config"},
			{Path: "internal/cmd/root_test.go", Code: "verification-surface-changed", Message: "changed test file"},
		},
	}
	assertReport(t, report, want)
}

func TestBuildReportRejectsInvalidChangedPaths(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{name: "empty", path: " ", wantErr: "changed path cannot be empty"},
		{name: "absolute", path: "/tmp/file.go", wantErr: "changed path must be repo-relative"},
		{name: "windows absolute", path: `C:\repo\internal\cmd\root.go`, wantErr: "changed path must be repo-relative"},
		{name: "parent traversal", path: "internal/../root.go", wantErr: "changed path cannot contain .."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildReport(Input{ChangedPaths: []string{tt.path}})
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestBuildReportRejectsInvalidScopeGlobs(t *testing.T) {
	_, err := BuildReport(Input{
		ChangedPaths: []string{"README.md"},
		Cards: []Card{{
			Path:  ".garden/context/broken.md",
			Scope: []string{"internal/[*.go"},
			Body:  "Broken guidance.",
		}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{
		".garden/context/broken.md",
		`invalid scope glob "internal/[*.go"`,
		"syntax error in pattern",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want substring %q", err.Error(), want)
		}
	}
}

func assertReport(t *testing.T, got Report, want Report) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("report = %#v, want %#v", got, want)
	}
}
