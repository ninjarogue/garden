package scopeglob

import (
	"path"
	"strings"
)

func Validate(pattern string) error {
	for _, patternPart := range strings.Split(pattern, "/") {
		if patternPart == "**" {
			continue
		}
		if _, err := path.Match(patternPart, ""); err != nil {
			return err
		}
	}
	return nil
}

func Match(pattern string, name string) (bool, error) {
	if err := Validate(pattern); err != nil {
		return false, err
	}
	patternParts := strings.Split(pattern, "/")
	nameParts := strings.Split(name, "/")
	return matchParts(patternParts, nameParts)
}

func matchParts(patternParts []string, nameParts []string) (bool, error) {
	if len(patternParts) == 0 {
		return len(nameParts) == 0, nil
	}
	if patternParts[0] == "**" {
		matched, err := matchParts(patternParts[1:], nameParts)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
		for i := range nameParts {
			matched, err := matchParts(patternParts[1:], nameParts[i+1:])
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	}
	if len(nameParts) == 0 {
		return false, nil
	}
	matched, err := path.Match(patternParts[0], nameParts[0])
	if err != nil {
		return false, err
	}
	if !matched {
		return false, nil
	}
	return matchParts(patternParts[1:], nameParts[1:])
}
