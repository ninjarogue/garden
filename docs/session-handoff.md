# Session Handoff

## 2026-05-25 15:28 CST

### Current State

Branch: `master`

Working tree was clean before this handoff update.

Recent commits:

- `39e91bb ref: simplify app sync flow`
- `b3c978e test: reduce command sync assertions`
- `9ba4919 fix: allow compact delimiters in card tags`
- `04cf0eb fix: lint missing AGENTS warning`
- `2d3d75e docs: remove garden edit handoff`
- `067ea97 fix: add generated AGENTS warning`

Latest verification before this handoff update:

```sh
garden lint
env GOCACHE=/tmp/garden-go-build go test ./...
```

Result: all packages passed.

The global binary at `/home/aric/.local/bin/garden` was rebuilt after behavior changes.

### Completed This Session

- Added a generated warning line inside Garden-managed `AGENTS.md` blocks.
- Synced this repo's `AGENTS.md` with global `garden agents sync --apply`.
- Added edge-case coverage so the warning stays on its own line even when existing marker content is inline.
- Added lint coverage for missing generated warnings.
- Relaxed tag validation: tags are human-only labels and may contain compact-index delimiter characters.
- Kept scope validation strict because scope is rendered into the AGENTS compact index.
- Reduced `internal/cmd` sync tests so command tests assert CLI UX and side effects instead of duplicating full generated output.
- Removed the `NewCardInput` alias and inlined the one-use app sync helper.
- Deleted `docs/garden-edit-decision-handoff.md`.

### Product Direction

Garden should continue proving the simple router workflow:

```txt
AGENTS.md = small always-visible router
.garden/context/*.md = human-readable context cards
garden = authoring, indexing, syncing, and linting tool
```

Do not add `garden edit`. Context cards should remain normal Markdown files. Edit cards directly, then run:

```sh
garden agents sync --apply
garden lint
```

### Architecture Notes

Current package direction is good:

- `internal/cmd`: Cobra command wiring and CLI UX.
- `internal/app`: use-case orchestration and app-owned DTOs/interfaces.
- `internal/contextcard`: Markdown card storage, parsing, validation, and template rendering.
- `internal/agents`: AGENTS marker/index rendering, sync behavior, and lint findings.
- `internal/output`: human-readable command output.

Avoid adding public packages, repository/usecase/entity layers, or other Clean Architecture ceremony. Keep future changes small and tied to a concrete dogfooding problem.

### What To Avoid

- Do not add new commands unless dogfooding exposes a clear workflow gap.
- Do not add subjective lint rules for vague tags, too many scopes, or secret-like words.
- Do not restrict tags based on AGENTS compact-index syntax while tags remain human-only.
- Do not duplicate full generated AGENTS output assertions in command tests; keep exact generated-output tests in the package that owns the output.

### Useful Verification

Run before handoff or commit:

```sh
garden lint
env GOCACHE=/tmp/garden-go-build go test ./...
```

If generated AGENTS output changes:

```sh
garden agents sync --apply
garden lint
```

### Next Steps

- Commit this log/handoff update if desired.
- Push `master` when ready.
- Continue dogfooding direct card edits plus sync/lint.

### Important Preferences

- Use TDD for behavior changes.
- Use Karpathy-guidelines for refactors and architecture cleanup.
- Keep changes surgical.
- Do not commit without explicit approval.
- Keep responses concise.
