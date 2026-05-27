# Session Handoff

## 2026-05-27

### Preliminary PR Review Handoff: Scope Glob Validation And CI Summary

Branch: `master`

Current branch state:

```txt
master includes merged PR #3 and is ahead of origin/master by local docs handoff cleanup commits.
```

What changed after reviewing PR #3:

- Added `internal/scopeglob` as the shared Garden glob helper.
- Moved Garden scope-glob syntax validation into context-card parsing/creation.
- Kept `internal/review` defensive so direct `review.BuildReport` callers still fail on invalid scope globs.
- Updated `garden lint` coverage so malformed scopes surface as `invalid-context-card`.
- Updated `.garden/context/context-card-format.md` and synced `AGENTS.md` for `internal/scopeglob/**`.
- Removed `docs/session-log.md`; `docs/session-handoff.md` is the single live handoff.

Important behavior:

- `scopeglob.Validate(pattern)` checks the whole Garden scope pattern before it is trusted.
- `scopeglob.Match(pattern, path)` now calls `Validate` first, so invalid later segments cannot hide behind an early non-match.
- Example fixed case: `scopeglob.Match("internal/[*.go", "README.md")` now returns a syntax error instead of `false, nil`.

Verification run after the fix:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
env GOCACHE=/tmp/garden-go-build go run ./cmd/garden lint
rg '"github.com/aric/garden/internal/(agents|contextcard|review|scopeglob)"' internal/cmd internal/output --glob '!**/*_test.go'
git diff --check
```

Result: all passed. A sub-agent review also found no current bugs; its residual risk about direct `scopeglob.Match` callers was fixed by validating inside `Match`.

Open product/design discussion for next session:

- For local `garden check`, current output keeps all matched cards under `Relevant constraints`, whether or not they contain `## Verification`.
- `Suggested verification` already filters to cards with extracted `## Verification`.
- For future PR CI summary output, one hypothesis is to prefer only cards with `## Verification` in the main evidence summary, or split non-verification cards into a smaller `Additional constraints` section.
- This is still open for discussion; do not treat it as a decided product requirement yet.
- No implementation has been started for this filtering yet.

### Current State

Branch: `master`

Branch has product-direction docs, the first `garden check` implementation slice, scope-glob validation, and the consolidated session handoff.

Notable changes:

- `docs/product-direction.md`
- `docs/check-command-implementation-handoff.md`
- `docs/changed-files-context-check-handoff.md`
- `docs/constraint-decay-garden-take.md`
- `docs/archive/product-direction-2026-05-26.md`
- `docs/archive/product-direction-2026-05-26-context-verification-thesis.md`
- `.garden/context/app-layer-architecture.md`
- `internal/review/**`
- `internal/app/**`
- `internal/cmd/**`
- `internal/output/**`
- `internal/scopeglob/**`
- `docs/agent-heavy-team-product-exploration.md`

Latest verification in this session:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
env GOCACHE=/tmp/garden-go-build go run ./cmd/garden lint
env GOCACHE=/tmp/garden-go-build go run ./cmd/garden check --changed go.mod --changed Makefile --changed .golangci.yml --changed internal/cmd/root.go --changed internal/cmd/root_test.go --changed docs/product-direction.md
rg '"github.com/aric/garden/internal/(agents|contextcard|review|scopeglob)"' internal/cmd internal/output --glob '!**/*_test.go'
git diff --check
```

Result: all passed. The sample `garden check` output included suggested verification from `.garden/context/app-layer-architecture.md` and flagged changed test, build config, and lint/format config files.

### Product Direction

Garden has pivoted from primarily an `AGENTS.md` router tool toward a PR context and verification reporter for agent-assisted software work.

Current thesis:

```txt
Garden turns changed files into reviewable constraint evidence.
```

At PR review time, Garden should answer:

```txt
What repo constraints apply to this change?
What evidence should prove they were preserved?
Did this PR also change the evidence?
```

Important framing:

- Rule files are advisory constraints.
- Tests and analysis are physical constraints.
- Garden should connect advisory context to reviewable verification evidence.
- Garden should not become an AI reviewer, policy engine, or arbitrary command runner.
- `AGENTS.md` remains the compact before-coding router.
- `.garden/context/*.md` cards are the shared data layer for both agent routing and PR review.

Primary doc:

- `docs/product-direction.md`

Historical archives:

- `docs/archive/product-direction-2026-05-26.md`
- `docs/archive/product-direction-2026-05-26-context-verification-thesis.md`

### Implementation Target

Implemented slice:

```sh
garden check --changed internal/cmd/root.go
```

This is a local preview command only. It is the engine for a future PR reporter and does not include GitHub Action integration yet.

Expected behavior:

```txt
changed paths
-> relevant context cards
-> suggested verification from cards
-> path-based verification-surface warnings
```

Detailed implementation handoff:

- `docs/check-command-implementation-handoff.md`

Older feature exploration:

- `docs/changed-files-context-check-handoff.md`

### Architecture Plan

Preserve the existing package boundary:

```txt
internal/cmd -> internal/app -> internal/review
internal/review -> internal/scopeglob
internal/output -> internal/app DTOs
```

Responsibilities:

- `internal/cmd`: Cobra wiring, `--changed` flag, CLI validation, command-level errors.
- `internal/app`: app-owned `CheckInput` / `CheckReport` DTOs, load cards through `CardStore`, call `internal/review`, adapt DTOs.
- `internal/review`: pure deterministic report logic.
- `internal/scopeglob`: Garden scope-glob validation and matching semantics.
- `internal/output`: human-readable `CheckReport` formatting.

Do not let `internal/cmd` call `internal/review` directly. Do not let `internal/output` import `internal/review`.

`internal/review` should stay narrow:

- Normalize changed paths.
- Match changed paths against card `scope` globs.
- Extract `## Verification` sections from card bodies.
- Detect verification surfaces from changed paths.
- Return deterministic report data.

`internal/scopeglob` should stay narrow:

- Validate Garden scope glob syntax.
- Match scope globs against slash-separated repo paths.
- Preserve `*` as one-segment matching and `**` as cross-directory matching.

### Key Decisions

- Start with repeated `--changed` flags only.
- Do not implement `--git-diff` yet.
- Do not implement `--changed-file-list` yet.
- Do not implement a GitHub Action yet.
- Do not execute verification commands.
- Do not infer trusted commands.
- Do not fail on uncovered files; show "no matching cards" as report state.
- Accept deleted-file-style paths without filesystem existence checks.
- Reject empty paths, absolute paths, and paths containing `..`.
- Normalize paths to slash-separated repo-relative paths.
- Use deterministic matching from card `scope` globs, not LLM judgment.
- Define `*` as not crossing `/`.
- Define `**` as crossing directories.
- Reject invalid scope glob syntax during context-card parse/create and defensively inside `review.BuildReport`.
- Sort changed paths, matched cards, and matched scopes for stable output.
- Detect verification surfaces path-only for v1: tests, `.github/workflows/**`, `.garden/context/**`, common build config files, and common lint/format config files.
- Defer section-level verification-change detection until a diff-aware mode exists.

### Test Plan

Follow:

- `.garden/context/testing-guidelines.md`
- `docs/testing.md`

Testing shape:

- `internal/review`: exact report structs for matching, verification extraction, warnings, ordering, no matches, and `**` semantics.
- `internal/scopeglob`: exact tests for `Validate`, `Match`, invalid syntax, `*`, and `**` behavior.
- `internal/app`: injected or temp-backed `CardStore` proving `App.Check` loads cards, adapts DTOs, and returns app-owned report data.
- `internal/output`: exact equality for stable `CheckReport` output.
- `internal/cmd`: CLI UX only; use temp dirs and real card files, assert key substrings rather than duplicating full output.

Run before handoff or commit:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
env GOCACHE=/tmp/garden-go-build go run ./cmd/garden lint
```

### Context Cards To Read Before Editing

Before editing implementation files:

- For `internal/app/**`, `internal/cmd/**`, or `internal/output/**`, read `.garden/context/app-layer-architecture.md`.
- For tests, read `.garden/context/testing-guidelines.md`.
- For context-card parsing or `internal/scopeglob/**` changes, read `.garden/context/context-card-format.md`.
- For AGENTS rendering changes, read `.garden/context/agents-router.md`.

The first `garden check` slice does not change AGENTS rendering logic. If context-card scopes change, run `garden agents sync --apply` so the generated index stays current.

### Maintenance Notes

- `docs/check-command-implementation-handoff.md` is historical implementation guidance for the completed first slice.
- `docs/changed-files-context-check-handoff.md` is now marked as superseded historical exploration.
- `docs/session-log.md` has been removed to avoid competing with this handoff.
- `README.md`, `.garden/context/product-direction.md`, `docs/testing.md`, and `.garden/context/testing-guidelines.md` have been updated for `garden check`.

### Important Preferences

- Use TDD for behavior changes.
- Keep changes surgical.
- Preserve the current package style; do not add Clean Architecture ceremony.
- Do not commit without explicit approval.
- Confirm docs updates with the user before making them.
