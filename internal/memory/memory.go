package memory

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"

	Version = 1
)

var idPattern = regexp.MustCompile(`^mem_[0-9a-f]{10}$`)

type Document struct {
	Version  int      `json:"version"`
	Memories []Memory `json:"memories"`
}

type Memory struct {
	ID        string   `json:"id"`
	Memory    string   `json:"memory"`
	Scope     []string `json:"scope"`
	Always    bool     `json:"always"`
	Tags      []string `json:"tags"`
	Priority  string   `json:"priority"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}

type Options struct {
	Scope    []string
	Always   bool
	Tags     []string
	Priority string
}

type IDGenerator func() (string, error)

func NewDocument() Document {
	return Document{Version: Version, Memories: []Memory{}}
}

func New(text string, opts Options, now time.Time, existingIDs map[string]bool, gen IDGenerator) (Memory, error) {
	text, opts, err := NormalizeOptions(text, opts)
	if err != nil {
		return Memory{}, err
	}
	if gen == nil {
		gen = RandomID
	}
	if existingIDs == nil {
		existingIDs = map[string]bool{}
	}

	var id string
	for {
		id, err = gen()
		if err != nil {
			return Memory{}, err
		}
		if !existingIDs[id] {
			break
		}
	}

	timestamp := FormatTimestamp(now)
	return Memory{
		ID:        id,
		Memory:    text,
		Scope:     opts.Scope,
		Always:    opts.Always,
		Tags:      opts.Tags,
		Priority:  opts.Priority,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}, nil
}

func NormalizeOptions(text string, opts Options) (string, Options, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", Options{}, fmt.Errorf("memory cannot be empty")
	}

	opts.Scope = cleanStrings(opts.Scope)
	opts.Tags = cleanStrings(opts.Tags)
	if opts.Priority == "" {
		opts.Priority = PriorityNormal
	}

	if len(opts.Scope) == 0 && !opts.Always {
		return "", Options{}, fmt.Errorf("use either --scope or --always")
	}
	if len(opts.Scope) > 0 && opts.Always {
		return "", Options{}, fmt.Errorf("--scope and --always cannot be used together")
	}
	if !ValidPriority(opts.Priority) {
		return "", Options{}, fmt.Errorf("invalid priority %q; expected low, normal, or high", opts.Priority)
	}

	return text, opts, nil
}

func Validate(mem Memory) error {
	if !ValidID(mem.ID) {
		return fmt.Errorf("invalid memory id %q", mem.ID)
	}
	_, _, err := NormalizeOptions(mem.Memory, Options{
		Scope:    mem.Scope,
		Always:   mem.Always,
		Tags:     mem.Tags,
		Priority: mem.Priority,
	})
	return err
}

func ExistingIDs(doc Document) map[string]bool {
	ids := make(map[string]bool, len(doc.Memories))
	for _, mem := range doc.Memories {
		ids[mem.ID] = true
	}
	return ids
}

func RandomID() (string, error) {
	buf := make([]byte, 5)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "mem_" + hex.EncodeToString(buf), nil
}

func ValidID(id string) bool {
	return idPattern.MatchString(id)
}

func ValidPriority(priority string) bool {
	switch priority {
	case PriorityLow, PriorityNormal, PriorityHigh:
		return true
	default:
		return false
	}
}

func FormatTimestamp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func ParseTimestamp(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	for _, layout := range []string{"2006-01-02T15:04:05.000Z", time.RFC3339Nano} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func NormalizeDocument(doc Document) Document {
	if doc.Version == 0 {
		doc.Version = Version
	}
	if doc.Memories == nil {
		doc.Memories = []Memory{}
	}
	for i := range doc.Memories {
		if doc.Memories[i].Scope == nil {
			doc.Memories[i].Scope = []string{}
		}
		if doc.Memories[i].Tags == nil {
			doc.Memories[i].Tags = []string{}
		}
		if doc.Memories[i].Priority == "" {
			doc.Memories[i].Priority = PriorityNormal
		}
	}
	return doc
}

func cleanStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}
	return cleaned
}
