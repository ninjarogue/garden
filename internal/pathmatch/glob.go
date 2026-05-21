package pathmatch

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func Any(patterns []string, target string) (bool, error) {
	for _, pattern := range patterns {
		matched, err := Match(pattern, target)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func Match(pattern string, target string) (bool, error) {
	return doublestar.PathMatch(normalize(pattern), normalize(target))
}

func normalize(value string) string {
	value = filepath.ToSlash(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "./")
	value = strings.TrimPrefix(value, "/")
	if value == "" {
		return "."
	}
	return path.Clean(value)
}
