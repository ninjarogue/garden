# Check Command Implementation Handoff

Status: implemented in PR #3. This file is retained as the implementation plan and historical handoff; use `docs/session-handoff.md` for current state.

This handoff covers the first implementation slice for `garden check`.

## Goal

Build a local preview command:

```sh
garden check --changed internal/cmd/root.go
```

The command should turn changed file paths into deterministic review context:

```txt
changed paths
-> relevant context cards
-> suggested verification from cards
-> path-based verification-surface warnings
```

This is the engine for the future PR reporter, but the first slice is local CLI only.

## Architecture

Preserve the existing package boundary:

```txt
internal/cmd -> internal/app -> internal/review
internal/output -> internal/app DTOs
```

Package responsibilities:

- `internal/cmd`: Cobra wiring, `--changed` flag, CLI validation, command-level errors.
- `internal/app`: app-owned `CheckInput` / `CheckReport` DTOs, load cards through `CardStore`, call `internal/review`, adapt DTOs.
- `internal/review`: pure deterministic report logic.
- `internal/scopeglob`: Garden scope-glob validation and matching semantics.
- `internal/output`: human-readable `CheckReport` formatting.

Do not let `internal/cmd` call `internal/review` directly. Do not let `internal/output` import `internal/review`.

## New Package: internal/review

Keep this package narrow. It should not know about Cobra, filesystems, `AGENTS.md`, or `internal/app`.

Suggested types:

```go
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
```

Responsibilities:

- Normalize changed paths to slash-separated repo-relative paths.
- Match changed paths against card `scope` globs.
- Extract the `## Verification` section from card bodies.
- Detect verification surfaces from paths.
- Return stable, deterministic report data.

## Path And Glob Decisions

Changed paths:

- Accept paths that do not exist on disk so deleted PR files still work.
- Reject empty paths.
- Reject absolute paths.
- Reject paths containing `..`.
- Normalize `./` and OS separators to slash-separated repo-relative paths.

Scope matching:

- Use card `scope` globs, not LLM judgment.
- The same changed files should always produce the same card list.
- `*` should not cross `/`.
- `**` should cross directories.
- Test cases should include `internal/cmd/**`, `.garden/context/**`, and `**/*_test.go`.

Matching semantics now live in `internal/scopeglob` so context-card validation and review matching use the same Garden glob rules.

## Verification Extraction

Start with a Markdown convention, not schema.

Extract the body of a `## Verification` section from `Card.Body`.

Rules:

- Header match should be exact enough to avoid surprising extraction.
- Stop at the next `## ` heading.
- Trim surrounding whitespace.
- If no verification section exists, leave suggested verification empty.

Do not infer trusted commands. Do not execute commands.

## Verification-Surface Warnings

First slice should be path-based only.

Flag changed paths that look like verification surfaces:

- `*_test.go`
- `.github/workflows/**`
- `.garden/context/**`
- common lint/format/build config files

With only `--changed` path input, Garden cannot know whether a specific `## Verification` section changed. Defer section-level change detection until a diff-aware mode exists.

A changed verification surface should be a report warning, not a command failure.

## Output Shape

The report should be compact and deterministic.

Recommended grouping:

```txt
Garden review context

Changed:
  internal/cmd/root.go

Relevant constraint:
  .garden/context/app-layer-architecture.md
  matched: internal/cmd/**

Suggested verification:
  ...

Verification surfaces changed:
  none
```

If a changed file has no matching cards, show that as report state. Do not fail the command.

Sort changed paths, matched cards, and matched scopes for stable output.

## First-Slice Non-Goals

Do not build yet:

- `--git-diff`
- `--changed-file-list`
- GitHub Action integration
- CI status behavior
- automatic execution of verification commands
- trusted-check allowlists
- blocking uncovered files
- AI relevance judgment
- card editing commands

## Test Plan

Follow `.garden/context/testing-guidelines.md`.

`internal/review`:

- One changed file matches one card.
- One changed file matches multiple cards.
- Multiple changed files match the same card.
- No matching card is represented without error.
- `**` matching works for nested paths.
- `**/*_test.go` matches test files.
- `## Verification` extraction works and stops at next `##`.
- Verification-surface paths produce warnings.
- Output report data is stable regardless of card input order.

`internal/scopeglob`:

- Invalid glob syntax returns an error.
- `**` crosses directories and can match zero segments.
- `*` stays within one path segment.

`internal/app`:

- Use injected or temp-backed `CardStore`.
- Prove `App.Check` loads cards, adapts DTOs, and returns app-owned report data.

`internal/output`:

- Exact equality for stable `CheckReport` output.
- Separate tests for matched cards, no matches, verification warnings, and suggested verification if needed.

`internal/cmd`:

- Test one CLI behavior at a time.
- `garden check --changed internal/cmd/root.go` succeeds with temp-dir real card files.
- Missing `--changed` returns an actionable error.
- Assert command UX and key substrings, not the full generated payload already covered by output tests.

## Verification

Run before handoff:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
env GOCACHE=/tmp/garden-go-build go run ./cmd/garden lint
```

## Resolved Decisions

- Use a fixed warning list for v1: tests, `.github/workflows/**`, `.garden/context/**`, common build config files, and common lint/format config files.
- Use repeated `--changed` flags first. Do not add positional changed paths in this slice.
- Group output by changed file first because reviewers start from the PR diff.
