package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aric/garden/internal/agents"
	"github.com/aric/garden/internal/memory"
	"github.com/aric/garden/internal/retrieval"
	"github.com/aric/garden/internal/storage"
)

type Options struct {
	Root        string
	Store       storage.Store
	Now         func() time.Time
	IDGenerator memory.IDGenerator
}

type App struct {
	root        string
	store       storage.Store
	now         func() time.Time
	idGenerator memory.IDGenerator
}

type RememberInput struct {
	Memory   string
	Scope    []string
	Always   bool
	Tags     []string
	Priority string
}

type PackInput struct {
	Path     string
	Task     string
	Max      int
	MaxChars int
}

type AgentsUpdateInput struct {
	Purpose    string
	Setup      []string
	Build      []string
	Lint       []string
	Typecheck  []string
	Test       []string
	Map        []agents.Entry
	Convention []string
	Doc        []agents.Entry
	Note       []string
	Apply      bool
}

const (
	agentsLintMaxLines = 250
	agentsLintMaxBytes = 12000
)

type AgentsSyncInput struct {
	Apply bool
}

type AgentsChange struct {
	Path     string
	Before   string
	After    string
	Applied  bool
	Findings []agents.Finding
}

type EditInput struct {
	ID          string
	Memory      string
	MemorySet   bool
	Scope       []string
	ScopeSet    bool
	Always      bool
	AlwaysSet   bool
	Tags        []string
	TagsSet     bool
	Priority    string
	PrioritySet bool
}

func New(opts Options) *App {
	if opts.Root == "" {
		opts.Root = "."
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.IDGenerator == nil {
		opts.IDGenerator = memory.RandomID
	}
	return &App{root: opts.Root, store: opts.Store, now: opts.Now, idGenerator: opts.IDGenerator}
}

func (a *App) Init() error {
	return a.store.Init()
}

func (a *App) Remember(input RememberInput) (memory.Memory, error) {
	opts := memory.Options{Scope: input.Scope, Always: input.Always, Tags: input.Tags, Priority: input.Priority}
	if _, _, err := memory.NormalizeOptions(input.Memory, opts); err != nil {
		return memory.Memory{}, err
	}

	doc, err := a.store.Load()
	if err != nil {
		return memory.Memory{}, err
	}
	mem, err := memory.New(input.Memory, opts, a.now(), memory.ExistingIDs(doc), a.idGenerator)
	if err != nil {
		return memory.Memory{}, err
	}
	doc.Memories = append(doc.Memories, mem)
	if err := a.store.Save(doc); err != nil {
		return memory.Memory{}, err
	}
	return mem, nil
}

func (a *App) Pack(input PackInput) ([]retrieval.Result, error) {
	if strings.TrimSpace(input.Path) == "" {
		return nil, fmt.Errorf("--path is required")
	}
	if strings.TrimSpace(input.Task) == "" {
		return nil, fmt.Errorf("--task is required")
	}
	if input.Max < 0 {
		return nil, fmt.Errorf("--max must be greater than 0")
	}
	if input.MaxChars < 0 {
		return nil, fmt.Errorf("--max-chars must be greater than 0")
	}

	doc, err := a.store.Load()
	if err != nil {
		return nil, err
	}
	return retrieval.Select(doc.Memories, retrieval.Query{
		Path:     input.Path,
		Task:     input.Task,
		Max:      input.Max,
		MaxChars: input.MaxChars,
	})
}

func (a *App) AgentsUpdate(input AgentsUpdateInput) (AgentsChange, error) {
	doc, err := a.store.Load()
	if err != nil {
		return AgentsChange{}, err
	}

	indexMemories := agentsIndexMemories(doc.Memories)
	block, err := agents.RenderBlock(agents.Context{
		Purpose:     input.Purpose,
		Setup:       input.Setup,
		Build:       input.Build,
		Lint:        input.Lint,
		Typecheck:   input.Typecheck,
		Test:        input.Test,
		Structure:   input.Map,
		Conventions: input.Convention,
		Docs:        input.Doc,
		Notes:       input.Note,
	}, indexMemories)
	if err != nil {
		return AgentsChange{}, err
	}

	return a.changeAgentsFile(input.Apply, func(current string) (string, error) {
		return agents.UpsertBlock(current, block)
	}, doc.Memories)
}

func (a *App) AgentsSync(input AgentsSyncInput) (AgentsChange, error) {
	doc, err := a.store.Load()
	if err != nil {
		return AgentsChange{}, err
	}
	indexMemories := agentsIndexMemories(doc.Memories)
	return a.changeAgentsFile(input.Apply, func(current string) (string, error) {
		return agents.SyncIndex(current, indexMemories)
	}, doc.Memories)
}

func (a *App) AgentsLint() ([]agents.Finding, error) {
	doc, err := a.store.Load()
	if err != nil {
		return nil, err
	}
	content, err := a.readAgentsFile()
	if err != nil {
		return nil, err
	}
	return agents.Lint(content, agentsLintOptions(doc.Memories)), nil
}

func (a *App) List() ([]memory.Memory, error) {
	doc, err := a.store.Load()
	if err != nil {
		return nil, err
	}
	return doc.Memories, nil
}

func (a *App) Remove(id string) error {
	id = strings.TrimSpace(id)
	doc, err := a.store.Load()
	if err != nil {
		return err
	}

	removed := false
	memories := make([]memory.Memory, 0, len(doc.Memories))
	for _, mem := range doc.Memories {
		if mem.ID == id {
			removed = true
			continue
		}
		memories = append(memories, mem)
	}
	if !removed {
		return fmt.Errorf("memory not found: %s", id)
	}
	doc.Memories = memories
	return a.store.Save(doc)
}

func (a *App) Edit(input EditInput) (memory.Memory, error) {
	input.ID = strings.TrimSpace(input.ID)
	if input.ID == "" {
		return memory.Memory{}, fmt.Errorf("memory id is required")
	}
	if !input.MemorySet && !input.ScopeSet && !input.AlwaysSet && !input.TagsSet && !input.PrioritySet {
		return memory.Memory{}, fmt.Errorf("use at least one edit flag")
	}
	if input.ScopeSet && input.AlwaysSet {
		return memory.Memory{}, fmt.Errorf("--scope and --always cannot be used together")
	}

	doc, err := a.store.Load()
	if err != nil {
		return memory.Memory{}, err
	}
	for i, mem := range doc.Memories {
		if mem.ID != input.ID {
			continue
		}

		if input.MemorySet {
			mem.Memory = input.Memory
		}
		if input.ScopeSet {
			mem.Scope = input.Scope
			mem.Always = false
		}
		if input.AlwaysSet {
			mem.Scope = []string{}
			mem.Always = input.Always
		}
		if input.TagsSet {
			mem.Tags = input.Tags
		}
		if input.PrioritySet {
			mem.Priority = input.Priority
		}

		_, opts, err := memory.NormalizeOptions(mem.Memory, memory.Options{
			Scope:    mem.Scope,
			Always:   mem.Always,
			Tags:     mem.Tags,
			Priority: mem.Priority,
		})
		if err != nil {
			return memory.Memory{}, err
		}
		mem.Scope = opts.Scope
		mem.Tags = opts.Tags
		mem.Priority = opts.Priority
		mem.UpdatedAt = memory.FormatTimestamp(a.now())

		doc.Memories[i] = mem
		if err := a.store.Save(doc); err != nil {
			return memory.Memory{}, err
		}
		return mem, nil
	}

	return memory.Memory{}, fmt.Errorf("memory not found: %s", input.ID)
}

func (a *App) changeAgentsFile(apply bool, render func(string) (string, error), memories []memory.Memory) (AgentsChange, error) {
	current, err := a.readAgentsFile()
	if err != nil {
		return AgentsChange{}, err
	}
	next, err := render(current)
	if err != nil {
		return AgentsChange{}, err
	}
	change := AgentsChange{
		Path:     a.agentsPath(),
		Before:   current,
		After:    next,
		Applied:  false,
		Findings: agents.Lint(next, agentsLintOptions(memories)),
	}
	if apply {
		if err := os.WriteFile(a.agentsPath(), []byte(next), 0o644); err != nil {
			return AgentsChange{}, err
		}
		change.Applied = true
	}
	return change, nil
}

func (a *App) readAgentsFile() (string, error) {
	data, err := os.ReadFile(a.agentsPath())
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *App) agentsPath() string {
	return filepath.Join(a.root, "AGENTS.md")
}

func memoryBodies(memories []memory.Memory) []string {
	bodies := make([]string, 0, len(memories))
	for _, mem := range memories {
		if strings.TrimSpace(mem.Memory) != "" {
			bodies = append(bodies, mem.Memory)
		}
	}
	return bodies
}

func agentsLintOptions(memories []memory.Memory) agents.LintOptions {
	return agents.LintOptions{
		MaxLines:     agentsLintMaxLines,
		MaxBytes:     agentsLintMaxBytes,
		MemoryBodies: memoryBodies(memories),
	}
}

func agentsIndexMemories(memories []memory.Memory) []agents.IndexMemory {
	indexMemories := make([]agents.IndexMemory, 0, len(memories))
	for _, mem := range memories {
		indexMemories = append(indexMemories, agents.IndexMemory{
			ID:     mem.ID,
			Scope:  mem.Scope,
			Always: mem.Always,
			Tags:   mem.Tags,
		})
	}
	return indexMemories
}
