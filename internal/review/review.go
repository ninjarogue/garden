package review

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/aric/garden/internal/scopeglob"
)

type Card struct {
	Path  string
	Scope []string
	Body  string
}

type Input struct {
	ChangedPaths []string
	Cards        []Card
}

type Report struct {
	ChangedFiles []ChangedFile
	Warnings     []Warning
}

type ChangedFile struct {
	Path  string
	Cards []MatchedCard
}

type MatchedCard struct {
	Path         string
	MatchedScope string
	Verification string
}

type Warning struct {
	Path    string
	Code    string
	Message string
}

func BuildReport(input Input) (Report, error) {
	changedPaths, err := normalizeChangedPaths(input.ChangedPaths)
	if err != nil {
		return Report{}, err
	}
	sort.Strings(changedPaths)

	cards := make([]Card, len(input.Cards))
	copy(cards, input.Cards)
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Path < cards[j].Path
	})
	if err := validateCardScopes(cards); err != nil {
		return Report{}, err
	}

	report := Report{ChangedFiles: make([]ChangedFile, 0, len(changedPaths))}
	for _, changedPath := range changedPaths {
		changedFile := ChangedFile{Path: changedPath}
		for _, card := range cards {
			matchedScopes, err := matchingScopes(changedPath, card.Scope)
			if err != nil {
				return Report{}, fmt.Errorf("%s: %w", card.Path, err)
			}
			for _, matchedScope := range matchedScopes {
				changedFile.Cards = append(changedFile.Cards, MatchedCard{
					Path:         card.Path,
					MatchedScope: matchedScope,
					Verification: extractVerification(card.Body),
				})
			}
		}
		report.ChangedFiles = append(report.ChangedFiles, changedFile)
		if warning, ok := verificationSurfaceWarning(changedPath); ok {
			report.Warnings = append(report.Warnings, warning)
		}
	}
	return report, nil
}

func validateCardScopes(cards []Card) error {
	for _, card := range cards {
		for _, scope := range card.Scope {
			if err := scopeglob.Validate(scope); err != nil {
				return fmt.Errorf("%s: invalid scope glob %q: %w", card.Path, scope, err)
			}
		}
	}
	return nil
}

func normalizeChangedPaths(paths []string) ([]string, error) {
	normalized := make([]string, 0, len(paths))
	for _, changedPath := range paths {
		changedPath = strings.TrimSpace(strings.ReplaceAll(changedPath, "\\", "/"))
		if changedPath == "" {
			return nil, fmt.Errorf("changed path cannot be empty")
		}
		if strings.HasPrefix(changedPath, "/") || hasWindowsDrivePrefix(changedPath) {
			return nil, fmt.Errorf("changed path must be repo-relative: %s", changedPath)
		}
		parts := strings.Split(changedPath, "/")
		for _, part := range parts {
			if part == ".." {
				return nil, fmt.Errorf("changed path cannot contain ..: %s", changedPath)
			}
		}
		changedPath = path.Clean(changedPath)
		changedPath = strings.TrimPrefix(changedPath, "./")
		normalized = append(normalized, changedPath)
	}
	return normalized, nil
}

func hasWindowsDrivePrefix(changedPath string) bool {
	return len(changedPath) >= 3 && changedPath[1] == ':' && changedPath[2] == '/'
}

func matchingScopes(changedPath string, scopes []string) ([]string, error) {
	matches := []string{}
	for _, scope := range scopes {
		matched, err := scopeglob.Match(scope, changedPath)
		if err != nil {
			return nil, fmt.Errorf("invalid scope glob %q: %w", scope, err)
		}
		if matched {
			matches = append(matches, scope)
		}
	}
	sort.Strings(matches)
	return matches, nil
}

func extractVerification(body string) string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	lines := strings.Split(body, "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## Verification" {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return ""
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines[start:end], "\n"))
}

func verificationSurfaceWarning(changedPath string) (Warning, bool) {
	switch {
	case strings.HasSuffix(changedPath, "_test.go"):
		return Warning{Path: changedPath, Code: "verification-surface-changed", Message: "changed test file"}, true
	case strings.HasPrefix(changedPath, ".github/workflows/"):
		return Warning{Path: changedPath, Code: "verification-surface-changed", Message: "changed GitHub workflow"}, true
	case strings.HasPrefix(changedPath, ".garden/context/"):
		return Warning{Path: changedPath, Code: "verification-surface-changed", Message: "changed Garden context card"}, true
	case isLintOrFormatConfig(changedPath):
		return Warning{Path: changedPath, Code: "verification-surface-changed", Message: "changed lint or format config"}, true
	case isBuildConfig(changedPath):
		return Warning{Path: changedPath, Code: "verification-surface-changed", Message: "changed build config"}, true
	default:
		return Warning{}, false
	}
}

func isLintOrFormatConfig(changedPath string) bool {
	base := path.Base(changedPath)
	switch base {
	case ".editorconfig", ".golangci.yml", ".golangci.yaml", ".prettierrc", ".prettierrc.json", ".prettierrc.yml", ".prettierrc.yaml", "biome.json":
		return true
	default:
		return strings.HasPrefix(base, "eslint.config.") || strings.HasPrefix(base, "prettier.config.")
	}
}

func isBuildConfig(changedPath string) bool {
	switch path.Base(changedPath) {
	case "go.mod", "go.sum", "Makefile", "makefile", "Dockerfile", "docker-compose.yml", "docker-compose.yaml", "package.json", "package-lock.json", "pnpm-lock.yaml", "yarn.lock", "Taskfile.yml", "Taskfile.yaml", "justfile":
		return true
	default:
		return false
	}
}
