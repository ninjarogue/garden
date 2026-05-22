# Session Handoff

## Current State

Branch: `master`

Working tree is intentionally dirty. Do not commit unless the user explicitly approves.

`docs/` is now unignored and should be treated as local project state going forward.

`README.md` has been replaced with the current Markdown context-card / AGENTS router command shape.

Current tests pass:

```txt
env GOCACHE=/tmp/garden-go-build go test ./...
```

## Product Direction

Garden should improve the Vercel-style `AGENTS.md` workflow.

Core model:

```txt
AGENTS.md = small always-visible router
.garden/context/*.md = human-readable context cards
garden = authoring, indexing, syncing, and linting tool
```

Do not center runtime context injection. `garden pack`, `garden read`, runtime ranking, session dedupe, and task-specific retrieval remain out of scope for now.

Agents should discover context through `AGENTS.md`, then use normal file-reading tools to inspect relevant Markdown cards.

## Implemented This Session

Core commands now exist in code:

```txt
garden init
garden new
garden remove
garden agents sync
garden lint
```

Key changes:

- Added `internal/contextcard` for Markdown card parsing/storage.
- `garden init` creates `.garden/context`.
- `garden new <slug>` creates `.garden/context/<slug>.md`.
- `garden remove <slug>` deletes a context card.
- `garden agents sync` renders a compact Garden-managed AGENTS block from cards.
- `garden lint` validates cards and checks AGENTS index freshness.
- `.gitignore` now allows `docs/` and `.garden/context/*.md`.
- `docs/compact-index-syntax.md` examples now use context cards.
- `README.md` now documents `init`, `new`, `agents sync`, `lint`, and `remove`.
- Legacy `internal/memory`, `internal/retrieval`, and `internal/storage` packages were removed.
- Follow-up review restored focused AGENTS marker/upsert tests and fixed YAML-safe card template rendering.
- Test suite cleanup split broad command/output tests, tightened stable generated-output assertions, and added focused coverage for RenderIndex validation, AGENTS lint marker cases, scalar scope parsing, and duplicate card slugs.

## Current AGENTS Index Shape

Use only the approved base compact syntax. Do not invent Garden-specific compact fields like `kind:`, `tags:`, or `card:`.

Current row shape:

```txt
[Garden Context Index]|root:.garden/context
|IMPORTANT:Before editing a listed area, inspect the matching context card
|src/routes/**:{rule,database,tenant-scoping,.garden/context/routes-query-modules.md}
```

The items inside `{...}` are plain compact index items:

- card kind
- tags
- card path

## Lint Scope

Current lint intent remains objective/basic:

- Each `.garden/context/*.md` file has YAML frontmatter.
- `kind` exists and is one of `rule`, `exception`, `warning`, `workflow`, or `background`.
- `scope` exists and has at least one non-empty glob.
- `scope` does not contain `CHANGE_ME`.
- `tags`, if present, is a list.
- Index metadata cannot contain compact-index delimiters that would make `AGENTS.md` unsyncable.
- Markdown body is non-empty.
- Card filename/slug is valid.
- `AGENTS.md` has the Garden managed block.
- `AGENTS.md` index matches current context cards.

Avoid subjective lint rules for now.

## Next Session Goal

Review the final diff, decide what should be committed, and avoid adding new product surface unless explicitly requested.

Suggested review areas:

1. Inspect the full dirty diff for accidental scope creep.
2. Decide whether `docs/session-*` should be committed or treated as local handoff notes.
3. If committing, keep the Markdown-card / AGENTS router changes separate from any unrelated future work.

## Known State/Risks

- `docs/` is now unignored, so existing docs are untracked unless added later.
- No commit has been made.
- TDD was used for the main behavior pivot; keep using tests before refactors that change behavior.

## Verification From This Session

Latest test run:

```txt
env GOCACHE=/tmp/garden-go-build go test ./...
```

Result: all packages pass.

## Important User Preferences

- Do not commit without explicit approval.
- Keep responses shorter.
- Start simple; avoid clever retrieval/session systems until the core AGENTS.md router workflow proves itself.
