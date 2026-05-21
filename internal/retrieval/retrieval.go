package retrieval

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/aric/garden/internal/memory"
	"github.com/aric/garden/internal/pathmatch"
)

const (
	DefaultMax      = 8
	DefaultMaxChars = 4000
)

type Query struct {
	Path     string
	Task     string
	Max      int
	MaxChars int
}

type Result struct {
	Memory  memory.Memory
	Score   int
	Reasons []Reason
}

type Reason struct {
	Text   string
	Points int
}

func Select(memories []memory.Memory, query Query) ([]Result, error) {
	if query.Max == 0 {
		query.Max = DefaultMax
	}
	if query.MaxChars == 0 {
		query.MaxChars = DefaultMaxChars
	}

	taskTokenList := Tokenize(query.Task)
	results := make([]Result, 0, len(memories))

	for _, mem := range memories {
		scopeMatched, matchedScope, err := firstMatchingScope(mem.Scope, query.Path)
		if err != nil {
			return nil, err
		}
		if !scopeMatched && !mem.Always {
			continue
		}

		score := 0
		reasons := []Reason{}
		if scopeMatched {
			score += 40
			reasons = append(reasons, Reason{Text: fmt.Sprintf("scope matched `%s`", matchedScope), Points: 40})
		}
		if mem.Always {
			score += 5
			reasons = append(reasons, Reason{Text: "always candidate", Points: 5})
		}

		memoryTokens := tokenSet(Tokenize(mem.Memory))
		for _, token := range taskTokenList {
			if memoryTokens[token] {
				score += 3
				reasons = append(reasons, Reason{Text: fmt.Sprintf("memory matched task token `%s`", token), Points: 3})
			}
		}

		tagTokens := map[string]bool{}
		tagLabels := map[string]string{}
		for _, tag := range mem.Tags {
			for _, token := range Tokenize(tag) {
				tagTokens[token] = true
				if tagLabels[token] == "" {
					tagLabels[token] = tag
				}
			}
		}
		for _, token := range taskTokenList {
			if tagTokens[token] {
				score += 8
				reasons = append(reasons, Reason{Text: fmt.Sprintf("tag `%s` matched task", tagLabels[token]), Points: 8})
			}
		}

		switch mem.Priority {
		case memory.PriorityHigh:
			score += 10
			reasons = append(reasons, Reason{Text: "priority high", Points: 10})
		case memory.PriorityLow:
			score -= 5
			reasons = append(reasons, Reason{Text: "priority low", Points: -5})
		}

		results = append(results, Result{Memory: mem, Score: score, Reasons: reasons})
	}

	sort.SliceStable(results, func(i, j int) bool {
		left := results[i]
		right := results[j]
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		leftUpdated := memory.ParseTimestamp(left.Memory.UpdatedAt)
		rightUpdated := memory.ParseTimestamp(right.Memory.UpdatedAt)
		if !leftUpdated.Equal(rightUpdated) {
			return leftUpdated.After(rightUpdated)
		}
		return left.Memory.ID < right.Memory.ID
	})

	return applyBudget(results, query.Max, query.MaxChars), nil
}

func firstMatchingScope(patterns []string, target string) (bool, string, error) {
	for _, pattern := range patterns {
		matched, err := pathmatch.Match(pattern, target)
		if err != nil {
			return false, "", err
		}
		if matched {
			return true, pattern, nil
		}
	}
	return false, "", nil
}

func Tokenize(text string) []string {
	tokens := []string{}
	seen := map[string]bool{}
	for _, field := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}) {
		if field == "" || seen[field] {
			continue
		}
		seen[field] = true
		tokens = append(tokens, field)
	}
	return tokens
}

func tokenSet(tokens []string) map[string]bool {
	set := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		set[token] = true
	}
	return set
}

func applyBudget(results []Result, max int, maxChars int) []Result {
	if max < 0 {
		max = 0
	}
	if maxChars < 0 {
		maxChars = 0
	}

	selected := make([]Result, 0, len(results))
	chars := 0
	for _, result := range results {
		if len(selected) >= max {
			break
		}
		memoryChars := len([]rune(result.Memory.Memory))
		if chars+memoryChars > maxChars {
			break
		}
		selected = append(selected, result)
		chars += memoryChars
	}
	return selected
}
