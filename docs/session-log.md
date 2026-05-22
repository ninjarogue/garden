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
- Compact AGENTS index should stay base Vercel-style syntax; no custom `kind:`, `tags:`, or `card:` fields.
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
