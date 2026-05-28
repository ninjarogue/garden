package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aric/garden/internal/agents"
	"github.com/aric/garden/internal/contextcard"
	"github.com/aric/garden/internal/review"
)

type Options struct {
	Root string
}

type App struct {
	cards      CardStore
	agentsFile AgentsFile
}

type Card struct {
	Slug  string
	Path  string
	Scope []string
	Tags  []string
	Body  string
}

type Finding struct {
	Severity string
	Code     string
	Message  string
}

type CheckInput struct {
	ChangedPaths []string
}

type CheckReport struct {
	ChangedFiles []CheckChangedFile
	Warnings     []CheckWarning
}

type CheckChangedFile struct {
	Path  string
	Cards []CheckMatchedCard
}

type CheckMatchedCard struct {
	Path         string
	MatchedScope string
	Verification string
}

type CheckWarning struct {
	Path    string
	Code    string
	Message string
}

type CreateCardInput struct {
	Slug  string
	Scope []string
	Tags  []string
}

type FileError struct {
	Path string
	Err  error
}

type CardStore interface {
	Init() error
	Create(CreateCardInput) (Card, error)
	Remove(slug string) (string, error)
	LoadAll() ([]Card, error)
	ReadAll() ([]Card, []FileError, error)
}

type AgentsFile interface {
	Read() (string, error)
	Write(content string) error
	Path() string
}

type AgentsSyncInput struct {
	Apply bool
}

type AgentsChange struct {
	Path     string
	Before   string
	After    string
	Applied  bool
	Findings []Finding
}

func New(opts Options) *App {
	if opts.Root == "" {
		opts.Root = "."
	}
	return &App{
		cards:      contextCardStore{store: contextcard.NewStore(opts.Root)},
		agentsFile: localAgentsFile{path: filepath.Join(opts.Root, "AGENTS.md")},
	}
}

func (a *App) Init() error {
	return a.cards.Init()
}

func (a *App) NewCard(input CreateCardInput) (Card, error) {
	return a.cards.Create(input)
}

func (a *App) RemoveCard(slug string) (string, error) {
	return a.cards.Remove(slug)
}

func (a *App) AgentsSync(input AgentsSyncInput) (AgentsChange, error) {
	cards, err := a.cards.LoadAll()
	if err != nil {
		return AgentsChange{}, err
	}
	indexCards := agentsIndexCards(cards)

	current, err := a.agentsFile.Read()
	if err != nil {
		return AgentsChange{}, err
	}
	next, err := agents.SyncIndex(current, indexCards)
	if err != nil {
		return AgentsChange{}, err
	}
	expected, err := agents.RenderIndex(indexCards)
	if err != nil {
		return AgentsChange{}, err
	}
	change := AgentsChange{
		Path:     a.agentsFile.Path(),
		Before:   current,
		After:    next,
		Applied:  false,
		Findings: appFindings(agents.Lint(next, agents.LintOptions{ExpectedIndex: expected})),
	}
	if input.Apply {
		if err := a.agentsFile.Write(next); err != nil {
			return AgentsChange{}, err
		}
		change.Applied = true
	}
	return change, nil
}

func (a *App) Lint() ([]Finding, error) {
	cards, fileErrors, err := a.cards.ReadAll()
	if err != nil {
		return nil, err
	}

	findings := make([]Finding, 0, len(fileErrors))
	for _, fileError := range fileErrors {
		findings = append(findings, Finding{
			Severity: "error",
			Code:     "invalid-context-card",
			Message:  fmt.Sprintf("%s: %v", fileError.Path, fileError.Err),
		})
	}

	expected, err := agents.RenderIndex(agentsIndexCards(cards))
	if err != nil {
		findings = append(findings, Finding{Severity: "error", Code: "invalid-garden-index", Message: err.Error()})
	} else {
		content, err := a.agentsFile.Read()
		if err != nil {
			return nil, err
		}
		findings = append(findings, appFindings(agents.Lint(content, agents.LintOptions{ExpectedIndex: expected}))...)
	}
	return findings, nil
}

func (a *App) Check(input CheckInput) (CheckReport, error) {
	cards, err := a.cards.LoadAll()
	if err != nil {
		return CheckReport{}, err
	}
	report, err := review.BuildReport(review.Input{
		ChangedPaths: input.ChangedPaths,
		Cards:        reviewCardsFromApp(cards),
	})
	if err != nil {
		return CheckReport{}, err
	}
	return checkReportFromReview(report), nil
}

type localAgentsFile struct {
	path string
}

func (f localAgentsFile) Read() (string, error) {
	data, err := os.ReadFile(f.path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f localAgentsFile) Write(content string) error {
	return os.WriteFile(f.path, []byte(content), 0o644)
}

func (f localAgentsFile) Path() string {
	return f.path
}

type contextCardStore struct {
	store *contextcard.Store
}

func (s contextCardStore) Init() error {
	return s.store.Init()
}

func (s contextCardStore) Create(input CreateCardInput) (Card, error) {
	card, err := s.store.Create(contextcard.CreateInput{
		Slug:  input.Slug,
		Scope: input.Scope,
		Tags:  input.Tags,
	})
	if err != nil {
		return Card{}, err
	}
	return appCard(card), nil
}

func (s contextCardStore) Remove(slug string) (string, error) {
	return s.store.Remove(slug)
}

func (s contextCardStore) LoadAll() ([]Card, error) {
	cards, err := s.store.LoadAll()
	if err != nil {
		return nil, err
	}
	return appCards(cards), nil
}

func (s contextCardStore) ReadAll() ([]Card, []FileError, error) {
	cards, fileErrors, err := s.store.ReadAll()
	return appCards(cards), appFileErrors(fileErrors), err
}

func agentsIndexCards(cards []Card) []agents.IndexCard {
	indexCards := make([]agents.IndexCard, 0, len(cards))
	for _, card := range cards {
		indexCards = append(indexCards, agents.IndexCard{
			Path:  card.Path,
			Scope: card.Scope,
		})
	}
	return indexCards
}

func appCards(cards []contextcard.Card) []Card {
	appCards := make([]Card, 0, len(cards))
	for _, card := range cards {
		appCards = append(appCards, appCard(card))
	}
	return appCards
}

func appCard(card contextcard.Card) Card {
	return Card{
		Slug:  card.Slug,
		Path:  card.Path,
		Scope: card.Scope,
		Tags:  card.Tags,
		Body:  card.Body,
	}
}

func appFileErrors(fileErrors []contextcard.FileError) []FileError {
	appErrors := make([]FileError, 0, len(fileErrors))
	for _, fileError := range fileErrors {
		appErrors = append(appErrors, FileError{
			Path: fileError.Path,
			Err:  fileError.Err,
		})
	}
	return appErrors
}

func appFindings(findings []agents.Finding) []Finding {
	appFindings := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		appFindings = append(appFindings, Finding{
			Severity: finding.Severity,
			Code:     finding.Code,
			Message:  finding.Message,
		})
	}
	return appFindings
}

func reviewCardsFromApp(cards []Card) []review.Card {
	reviewCards := make([]review.Card, 0, len(cards))
	for _, card := range cards {
		reviewCards = append(reviewCards, review.Card{
			Path:  card.Path,
			Scope: card.Scope,
			Body:  card.Body,
		})
	}
	return reviewCards
}

func checkReportFromReview(report review.Report) CheckReport {
	return CheckReport{
		ChangedFiles: checkChangedFilesFromReview(report.ChangedFiles),
		Warnings:     checkWarningsFromReview(report.Warnings),
	}
}

func checkChangedFilesFromReview(changedFiles []review.ChangedFile) []CheckChangedFile {
	appFiles := make([]CheckChangedFile, 0, len(changedFiles))
	for _, changedFile := range changedFiles {
		appFiles = append(appFiles, CheckChangedFile{
			Path:  changedFile.Path,
			Cards: checkMatchedCardsFromReview(changedFile.Cards),
		})
	}
	return appFiles
}

func checkMatchedCardsFromReview(cards []review.MatchedCard) []CheckMatchedCard {
	appCards := make([]CheckMatchedCard, 0, len(cards))
	for _, card := range cards {
		appCards = append(appCards, CheckMatchedCard{
			Path:         card.Path,
			MatchedScope: card.MatchedScope,
			Verification: card.Verification,
		})
	}
	return appCards
}

func checkWarningsFromReview(warnings []review.Warning) []CheckWarning {
	appWarnings := make([]CheckWarning, 0, len(warnings))
	for _, warning := range warnings {
		appWarnings = append(appWarnings, CheckWarning{
			Path:    warning.Path,
			Code:    warning.Code,
			Message: warning.Message,
		})
	}
	return appWarnings
}
