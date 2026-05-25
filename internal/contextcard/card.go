package contextcard

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Card struct {
	Slug  string
	Path  string
	Scope []string
	Tags  []string
	Body  string
}

type CreateInput struct {
	Slug  string
	Scope []string
	Tags  []string
}

type FileError struct {
	Path string
	Err  error
}

type Store struct {
	root string
}

func NewStore(root string) *Store {
	if root == "" {
		root = "."
	}
	return &Store{root: root}
}

func (s *Store) Init() error {
	return os.MkdirAll(s.contextDir(), 0o755)
}

func (s *Store) Create(input CreateInput) (Card, error) {
	input.Slug = strings.TrimSpace(input.Slug)
	if err := validateSlug(input.Slug); err != nil {
		return Card{}, err
	}
	scope := cleanStrings(input.Scope)
	if len(scope) == 0 {
		return Card{}, fmt.Errorf("scope must include at least one glob")
	}
	for _, value := range scope {
		if strings.Contains(value, "CHANGE_ME") {
			return Card{}, fmt.Errorf("scope cannot contain CHANGE_ME")
		}
	}
	tags := cleanStrings(input.Tags)
	if err := validateCompactIndexMetadata(scope, tags); err != nil {
		return Card{}, err
	}

	relPath := cardPath(input.Slug)
	absPath := filepath.Join(s.root, filepath.FromSlash(relPath))
	content := renderTemplate(input.Slug, scope, tags)
	card, err := Parse(relPath, []byte(content))
	if err != nil {
		return Card{}, err
	}

	if err := s.Init(); err != nil {
		return Card{}, err
	}
	file, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return Card{}, fmt.Errorf("context card already exists: %s", relPath)
	}
	if err != nil {
		return Card{}, err
	}
	defer file.Close()
	if _, err := file.WriteString(content); err != nil {
		return Card{}, err
	}
	return card, nil
}

func (s *Store) Remove(slug string) (string, error) {
	slug = strings.TrimSpace(slug)
	if err := validateSlug(slug); err != nil {
		return "", err
	}
	relPath := cardPath(slug)
	err := os.Remove(filepath.Join(s.root, filepath.FromSlash(relPath)))
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("context card not found: %s", relPath)
	}
	if err != nil {
		return "", err
	}
	return relPath, nil
}

func (s *Store) LoadAll() ([]Card, error) {
	cards, fileErrors, err := s.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(fileErrors) > 0 {
		return nil, fmt.Errorf("invalid context card %s: %w", fileErrors[0].Path, fileErrors[0].Err)
	}
	return cards, nil
}

func (s *Store) ReadAll() ([]Card, []FileError, error) {
	matches, err := filepath.Glob(filepath.Join(s.contextDir(), "*.md"))
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(matches)

	cards := make([]Card, 0, len(matches))
	fileErrors := []FileError{}
	for _, match := range matches {
		data, err := os.ReadFile(match)
		relPath := filepath.ToSlash(filepath.Join(".garden", "context", filepath.Base(match)))
		if err != nil {
			return nil, nil, err
		}
		card, err := Parse(relPath, data)
		if err != nil {
			fileErrors = append(fileErrors, FileError{Path: relPath, Err: err})
			continue
		}
		cards = append(cards, card)
	}
	return cards, fileErrors, nil
}

func Parse(path string, data []byte) (Card, error) {
	path = filepath.ToSlash(strings.TrimSpace(path))
	slug := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if err := validateSlug(slug); err != nil {
		return Card{}, err
	}

	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(content, "---\n") {
		return Card{}, fmt.Errorf("YAML frontmatter is required")
	}
	rest := content[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return Card{}, fmt.Errorf("YAML frontmatter is required")
	}
	frontmatter := rest[:end]
	body := strings.TrimSpace(strings.TrimPrefix(rest[end+len("\n---"):], "\n"))

	var meta struct {
		Scope []string `yaml:"scope"`
		Tags  []string `yaml:"tags"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		if strings.Contains(err.Error(), "tags") || fieldLooksScalar(frontmatter, "tags") {
			return Card{}, fmt.Errorf("tags must be a list")
		}
		if strings.Contains(err.Error(), "scope") || fieldLooksScalar(frontmatter, "scope") {
			return Card{}, fmt.Errorf("scope must be a list")
		}
		return Card{}, fmt.Errorf("read YAML frontmatter: %w", err)
	}

	scope := cleanStrings(meta.Scope)
	if len(scope) == 0 {
		return Card{}, fmt.Errorf("scope must include at least one glob")
	}
	for _, value := range scope {
		if strings.Contains(value, "CHANGE_ME") {
			return Card{}, fmt.Errorf("scope cannot contain CHANGE_ME")
		}
	}
	tags := cleanStrings(meta.Tags)
	if err := validateCompactIndexMetadata(scope, tags); err != nil {
		return Card{}, err
	}
	if body == "" {
		return Card{}, fmt.Errorf("body cannot be empty")
	}

	return Card{
		Slug:  slug,
		Path:  path,
		Scope: scope,
		Tags:  tags,
		Body:  body,
	}, nil
}

func validateSlug(slug string) error {
	if !slugPattern.MatchString(strings.TrimSpace(slug)) {
		return fmt.Errorf("invalid card slug %q; expected lowercase words separated by hyphens", slug)
	}
	return nil
}

func renderTemplate(slug string, scope []string, tags []string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("scope:\n")
	for _, value := range scope {
		b.WriteString("  - ")
		b.WriteString(yamlScalar(value))
		b.WriteString("\n")
	}
	if len(tags) > 0 {
		b.WriteString("tags:\n")
		for _, value := range tags {
			b.WriteString("  - ")
			b.WriteString(yamlScalar(value))
			b.WriteString("\n")
		}
	}
	b.WriteString("---\n\n")
	b.WriteString("# ")
	b.WriteString(titleFromSlug(slug))
	b.WriteString("\n\n")
	b.WriteString("Write the repo context here.\n")
	return b.String()
}

func cardPath(slug string) string {
	return filepath.ToSlash(filepath.Join(".garden", "context", slug+".md"))
}

func (s *Store) contextDir() string {
	return filepath.Join(s.root, ".garden", "context")
}

func titleFromSlug(slug string) string {
	parts := strings.Split(slug, "-")
	for i, part := range parts {
		runes := []rune(part)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}

func cleanStrings(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}
	return cleaned
}

func validateCompactIndexMetadata(scope []string, tags []string) error {
	for _, value := range scope {
		if err := rejectCompactIndexRowSyntax("scope", value); err != nil {
			return err
		}
	}
	for _, value := range tags {
		if err := rejectCompactIndexItemSyntax("tag", value); err != nil {
			return err
		}
	}
	return nil
}

func rejectCompactIndexRowSyntax(field string, value string) error {
	for _, r := range value {
		switch r {
		case '|':
			return fmt.Errorf("%s contains compact index syntax delimiter %q", field, r)
		}
		if unicode.IsControl(r) {
			return fmt.Errorf("%s contains compact index syntax control character", field)
		}
	}
	return nil
}

func rejectCompactIndexItemSyntax(field string, value string) error {
	for _, r := range value {
		switch r {
		case '|', '{', '}', ',':
			return fmt.Errorf("%s contains compact index syntax delimiter %q", field, r)
		}
		if unicode.IsControl(r) {
			return fmt.Errorf("%s contains compact index syntax control character", field)
		}
	}
	return nil
}

func yamlScalar(value string) string {
	data, err := yaml.Marshal(value)
	if err != nil {
		return value
	}
	return strings.TrimSpace(string(data))
}

func fieldLooksScalar(frontmatter string, field string) bool {
	prefix := field + ":"
	for _, line := range strings.Split(frontmatter, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) && strings.TrimSpace(strings.TrimPrefix(line, prefix)) != "" {
			return true
		}
	}
	return false
}
