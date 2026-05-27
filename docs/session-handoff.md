# Session Handoff

## 2026-05-27

### Current State

Branch: `feature/garden-check`

Branch has product-direction docs and the first `garden check` implementation slice.

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
- `docs/agent-heavy-team-product-exploration.md`

Latest verification in this session:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
env GOCACHE=/tmp/garden-go-build go run ./cmd/garden lint
env GOCACHE=/tmp/garden-go-build go run ./cmd/garden check --changed go.mod --changed Makefile --changed .golangci.yml --changed internal/cmd/root.go --changed internal/cmd/root_test.go --changed docs/product-direction.md
rg '"github.com/aric/garden/internal/(agents|contextcard|review)"' internal/cmd internal/output --glob '!**/*_test.go'
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
internal/output -> internal/app DTOs
```

Responsibilities:

- `internal/cmd`: Cobra wiring, `--changed` flag, CLI validation, command-level errors.
- `internal/app`: app-owned `CheckInput` / `CheckReport` DTOs, load cards through `CardStore`, call `internal/review`, adapt DTOs.
- `internal/review`: pure deterministic report logic.
- `internal/output`: human-readable `CheckReport` formatting.

Do not let `internal/cmd` call `internal/review` directly. Do not let `internal/output` import `internal/review`.

`internal/review` should stay narrow:

- Normalize changed paths.
- Match changed paths against card `scope` globs.
- Extract `## Verification` sections from card bodies.
- Detect verification surfaces from changed paths.
- Return deterministic report data.

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
- Sort changed paths, matched cards, and matched scopes for stable output.
- Detect verification surfaces path-only for v1: tests, `.github/workflows/**`, `.garden/context/**`, common build config files, and common lint/format config files.
- Defer section-level verification-change detection until a diff-aware mode exists.

### Test Plan

Follow:

- `.garden/context/testing-guidelines.md`
- `docs/testing.md`

Testing shape:

- `internal/review`: exact report structs for matching, verification extraction, warnings, ordering, no matches, and `**` semantics.
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
- For context-card parsing changes, read `.garden/context/context-card-format.md`.
- For AGENTS rendering changes, read `.garden/context/agents-router.md`.

The first `garden check` slice should not require AGENTS rendering changes.

### Maintenance Notes

- `docs/check-command-implementation-handoff.md` is the active implementation handoff for this slice.
- `docs/changed-files-context-check-handoff.md` is now marked as superseded historical exploration.
- `README.md`, `.garden/context/product-direction.md`, `docs/testing.md`, and `.garden/context/testing-guidelines.md` have been updated for `garden check`.

### Important Preferences

- Use TDD for behavior changes.
- Keep changes surgical.
- Preserve the current package style; do not add Clean Architecture ceremony.
- Do not commit without explicit approval.
- Confirm docs updates with the user before making them.
