package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aric/garden/internal/agents"
	"github.com/aric/garden/internal/contextcard"
)

type Options struct {
	Root string
}

type App struct {
	root  string
	cards *contextcard.Store
}

type NewCardInput struct {
	Slug  string
	Kind  string
	Scope []string
	Tags  []string
}

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

func New(opts Options) *App {
	if opts.Root == "" {
		opts.Root = "."
	}
	return &App{root: opts.Root, cards: contextcard.NewStore(opts.Root)}
}

func (a *App) Init() error {
	return a.cards.Init()
}

func (a *App) NewCard(input NewCardInput) (contextcard.Card, error) {
	return a.cards.Create(contextcard.CreateInput{
		Slug:  input.Slug,
		Kind:  input.Kind,
		Scope: input.Scope,
		Tags:  input.Tags,
	})
}

func (a *App) RemoveCard(slug string) (string, error) {
	return a.cards.Remove(slug)
}

func (a *App) ListCards() ([]contextcard.Card, error) {
	return a.cards.LoadAll()
}

func (a *App) AgentsSync(input AgentsSyncInput) (AgentsChange, error) {
	cards, err := a.cards.LoadAll()
	if err != nil {
		return AgentsChange{}, err
	}
	indexCards := agentsIndexCards(cards)
	return a.changeAgentsFile(input.Apply, func(current string) (string, error) {
		return agents.SyncIndex(current, indexCards)
	}, indexCards)
}

func (a *App) Lint() ([]agents.Finding, error) {
	cards, fileErrors, err := a.cards.ReadAll()
	if err != nil {
		return nil, err
	}

	findings := make([]agents.Finding, 0, len(fileErrors))
	for _, fileError := range fileErrors {
		findings = append(findings, agents.Finding{
			Severity: "error",
			Code:     "invalid-context-card",
			Message:  fmt.Sprintf("%s: %v", fileError.Path, fileError.Err),
		})
	}

	expected, err := agents.RenderIndex(agentsIndexCards(cards))
	if err != nil {
		findings = append(findings, agents.Finding{Severity: "error", Code: "invalid-garden-index", Message: err.Error()})
	} else {
		content, err := a.readAgentsFile()
		if err != nil {
			return nil, err
		}
		findings = append(findings, agents.Lint(content, agents.LintOptions{ExpectedIndex: expected})...)
	}
	return findings, nil
}

func (a *App) changeAgentsFile(apply bool, render func(string) (string, error), indexCards []agents.IndexCard) (AgentsChange, error) {
	current, err := a.readAgentsFile()
	if err != nil {
		return AgentsChange{}, err
	}
	next, err := render(current)
	if err != nil {
		return AgentsChange{}, err
	}
	expected, err := agents.RenderIndex(indexCards)
	if err != nil {
		return AgentsChange{}, err
	}
	change := AgentsChange{
		Path:     a.agentsPath(),
		Before:   current,
		After:    next,
		Applied:  false,
		Findings: agents.Lint(next, agents.LintOptions{ExpectedIndex: expected}),
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

func agentsIndexCards(cards []contextcard.Card) []agents.IndexCard {
	indexCards := make([]agents.IndexCard, 0, len(cards))
	for _, card := range cards {
		indexCards = append(indexCards, agents.IndexCard{
			Path:  card.Path,
			Kind:  card.Kind,
			Scope: card.Scope,
			Tags:  card.Tags,
		})
	}
	return indexCards
}
