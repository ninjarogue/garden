package scopeglob

import (
	"strings"
	"testing"
)

func TestValidateRejectsInvalidGlobSyntax(t *testing.T) {
	err := Validate("internal/[*.go")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "syntax error in pattern") {
		t.Fatalf("error = %q, want syntax error", err.Error())
	}
}

func TestMatchHandlesGardenDoubleStar(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{name: "nested path", pattern: "internal/cmd/**", path: "internal/cmd/sub/root.go", want: true},
		{name: "zero segments", pattern: "internal/cmd/**", path: "internal/cmd", want: true},
		{name: "test file anywhere", pattern: "**/*_test.go", path: "src/deep/root_test.go", want: true},
		{name: "single star stays in segment", pattern: "internal/*.go", path: "internal/cmd/root.go", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Match(tt.pattern, tt.path)
			if err != nil {
				t.Fatalf("Match returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Match(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestMatchReturnsInvalidGlobError(t *testing.T) {
	_, err := Match("internal/[*.go", "internal/root.go")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "syntax error in pattern") {
		t.Fatalf("error = %q, want syntax error", err.Error())
	}
}

func TestMatchValidatesFullPatternBeforeMatching(t *testing.T) {
	_, err := Match("internal/[*.go", "README.md")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "syntax error in pattern") {
		t.Fatalf("error = %q, want syntax error", err.Error())
	}
}
