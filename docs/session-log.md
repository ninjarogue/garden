# Session Log

## 2026-05-22 11:00 CST

### Context
- Repo: `/home/aric/dev/garden`
- Branch: `master`
- Constraint: do not commit without explicit approval.
- Direction: `AGENTS.md` is the router; `.garden/context/*.md` cards hold detailed context.

### Executive Summary
- Started moving Garden from JSON memories/runtime retrieval toward Markdown context cards and AGENTS router sync.
- Used TDD: wrote failing tests first, then implemented the smallest passing slice.

### Findings
- `.gitignore` ignored both `docs/` and `.garden/`; `docs/` is now trackable, and `.garden/context/*.md` is allowed.
- Compact AGENTS index should stay base Vercel-style syntax; no custom row-level fields.
- Legacy memory/retrieval/storage packages remain present but are no longer wired into app/CLI.

### Progress Completed
- Added `internal/contextcard` parser/store with YAML frontmatter validation.
- Reworked CLI to core commands: `init`, `new`, `remove`, `agents sync`, `lint`.
- Reworked AGENTS sync to emit `[Garden Context Index]|root:.garden/context`.
- Added linting for invalid cards and stale/missing AGENTS index.
- Updated `docs/compact-index-syntax.md` examples for context cards.
- Verified: `env GOCACHE=/tmp/garden-go-build mise exec -- go test ./...`.

### Open Blockers/Risks
- `README.md` remains deleted from prior state.
- `docs/` is newly unignored, so existing docs are now untracked.
- Legacy `internal/memory`, `internal/retrieval`, and `internal/storage` still needed cleanup at this point.

### Next Steps
- Delete or archive disconnected legacy memory/retrieval/storage surface.
- Add/refresh tracked README or docs for the new product direction.
- Manually inspect generated card and AGENTS output with the CLI.

### Tags
#codex #session-log #progress #garden

## 2026-05-22 11:05 CST

### Context
- Follow-up cleanup after first Markdown card/AGENTS router slice.

### Executive Summary
- Removed disconnected legacy runtime retrieval surface.

### Findings
- App/CLI had no live imports of `internal/memory`, `internal/retrieval`, or `internal/storage`.
- Empty package directories still existed after file deletion; removed them too.

### Progress Completed
- Used a temporary failing guard test to drive legacy package removal, then removed it.
- Deleted `internal/memory`, `internal/retrieval`, and `internal/storage`.
- Verified: `env GOCACHE=/tmp/garden-go-build mise exec -- go test ./...`.

### Open Blockers/Risks
- `README.md` remains deleted from prior state.
- Docs are now trackable but still uncommitted.

### Next Steps
- Refresh README or tracked docs for the new core command shape.
- Manually inspect generated CLI output in a temp repo.

### Tags
#codex #session-log #progress #garden

## 2026-05-22 12:10 CST

### Context
- Follow-up review/refactor pass for Markdown context cards and AGENTS router.
- Constraint: do not commit without explicit approval.

### Executive Summary
- Reviewed package boundaries and CLI behavior after the product pivot.
- Restored focused AGENTS marker/upsert tests.
- Fixed real bugs found during review around YAML-sensitive glob scopes and unsyncable compact-index metadata.

### Findings
- `garden new --scope '**/*'` failed because generated YAML wrote raw `**/*`; it also left the invalid file behind.
- Tags containing compact-index delimiters, such as `database,tenant`, could create a card that later failed `garden agents sync`.
- `README.md` and product docs needed to reflect the new command shape and compact index row syntax.

### Progress Completed
- `contextcard.Store.Create` now renders YAML-safe scalar values and parses the generated card before writing.
- Card parsing/creation rejects metadata that would break the generated compact index.
- Added focused AGENTS marker/upsert tests and fixed extra blank-line output when refreshing an existing index.
- Replaced the stale README with current usage for `init`, `new`, `agents sync`, `lint`, and `remove`.
- Verified with `env GOCACHE=/tmp/garden-go-build go test ./...` and manual CLI smoke tests.

### Open Blockers/Risks
- Working tree remains intentionally dirty and uncommitted.
- `docs/` is now trackable but still needs a commit/no-commit decision.

### Next Steps
- Review the final dirty diff before committing.
- Keep future changes focused on the AGENTS router/context-card workflow unless the product direction changes.

### Tags
#codex #session-log #review #garden

## 2026-05-22 12:25 CST

### Context
- Test-suite cleanup after the implementation review.

### Executive Summary
- Improved tests without changing production behavior.
- Split broad tests into focused behavior checks and tightened stable generated-output assertions.

### Findings
- `internal/cmd/root_test.go` had one broad workflow test covering init, new, sync, and lint.
- `internal/output/output_test.go` combined preview diff output, lint findings, no-change output, and applied messaging.
- Several generated-output tests used substring checks where exact equality gives better diagnostics.

### Progress Completed
- Split CLI tests into focused command behaviors.
- Split output formatting tests into focused preview, findings, applied/no-change, lint pass, and lint findings cases.
- Added direct `agents.RenderIndex` validation coverage for reserved markers, compact-index delimiters, and control characters.
- Added separate `agents.Lint` tests for missing Garden agents block, missing Garden index, malformed Garden index markers, and stale index.
- Added `contextcard.Parse` scalar scope coverage and `contextcard.Store.Create` duplicate slug coverage.
- Kept YAML-sensitive glob and unsyncable-tag tests intact.
- Verified with `env GOCACHE=/tmp/garden-go-build go test ./...`.

### Open Blockers/Risks
- No production behavior changed in this test cleanup.
- Dirty tree still includes the larger product pivot and docs updates.

### Next Steps
- Review and commit only after explicit approval.

### Tags
#codex #session-log #tests #garden

## 2026-05-25 09:09 CST

### Context
- Follow-up app-layer refactor session.
- User requested TDD and Karpathy-guidelines for a behavior-preserving app-layer refactor.

### Executive Summary
- Decoupled `internal/app` so command/output code talks to app-owned use-case types.
- Committed the refactor as `1c42150 ref: decouple app layer types`.

### Findings
- `internal/app` previously exposed lower-level `contextcard.Card` and `agents.Finding` types.
- `internal/output` also depended directly on `agents.Finding`.
- App filesystem access for `AGENTS.md` was embedded in app methods, which made isolated app tests harder.

### Progress Completed
- Added app-owned `Card`, `Finding`, `CreateCardInput`, `FileError`, `CardStore`, and `AgentsFile` types.
- Added adapters for `contextcard.Store` and local `AGENTS.md` file access.
- Updated `cmd` and `output` to consume `app.AgentsChange` and `app.Finding`.
- Added focused tests for injected card storage, injected `AGENTS.md` access, and app-owned findings.
- Verified with `env GOCACHE=/tmp/garden-go-build go test ./...`.

### Open Blockers/Risks
- Branch is ahead of `origin/master` by one commit.
- `docs/session-handoff.md` should be disregarded until the next product slice is clarified.

### Next Steps
- Review product docs and decide the next Garden product slice before adding more surface area.
- Push committed refactor work when ready.

### Tags
#codex #session-log #refactor #garden

## 2026-05-25 15:28 CST

### Context
- Follow-up dogfooding and architecture review session on `master`.
- User asked to carry out the Garden edit handoff, review the codebase with an architecture lens, and keep changes surgical.
- Constraint: do not commit without explicit approval.

### Executive Summary
- Added a generated warning inside Garden-managed `AGENTS.md` blocks and made lint protect that warning.
- Reversed the over-coupled tag validation rule so tags are human-only labels again.
- Reduced duplicated command-test assertions and simplified a small app-layer indirection.
- Committed the completed changes in focused commits.

### Findings
- A generated warning in `AGENTS.md` is a useful low-cost guardrail against hand-editing the managed block.
- `garden lint` needed to enforce that guardrail; otherwise a stale block with a current index could pass.
- Compact-index delimiter restrictions belonged on `scope`, not on human-only `tags`.
- Command tests were repeating full generated AGENTS output already owned by lower-level tests.
- `NewCardInput` and `changeAgentsFile` added names and hops without enough current value.

### Progress Completed
- Added generated warning support and synced `AGENTS.md`.
- Fixed warning insertion for inline marker/content edge cases.
- Added `missing-generated-warning` lint coverage.
- Allowed compact-index delimiter characters in tags while keeping scope validation.
- Updated context/product/PBT/testing docs to match current behavior.
- Reduced `internal/cmd` sync tests to command UX and side-effect assertions.
- Removed `NewCardInput` alias and inlined the one-use app sync helper.
- Rebuilt the global `/home/aric/.local/bin/garden` binary after behavior changes.
- Verified with `garden lint` and `env GOCACHE=/tmp/garden-go-build go test ./...`.

### Open Blockers/Risks
- No known blockers.
- Further architecture cleanup should stay opportunistic; avoid adding layers or product surface.

### Next Steps
- Refresh this handoff and commit log docs if desired.
- Push current `master` when ready.
- Continue dogfooding direct Markdown card edits plus `garden agents sync --apply` and `garden lint`.

### Tags
#codex #session-log #architecture #garden
