# Changed Files Context Check Handoff

Status: historical feature exploration. The feature is now implemented as `garden check <changed-path>...`; early command sketches below are not current CLI syntax.

This handoff captures a future Garden feature: given a set of changed files, deterministically report the relevant Garden context cards.

Related exploration:

- `docs/constraint-decay-garden-take.md`

## Feature Goal

Garden should answer:

```txt
Given these changed files, which Garden context cards should humans, agents, or CI look at?
```

This should be deterministic. Garden should use context-card `scope` globs, not LLM judgment, to map changed files to cards.

## Why This Matters

Garden currently helps agents discover context through the `AGENTS.md` router. That is useful, but advisory context alone does not prove the relevant rules were considered during a change.

This feature would make scoped context useful in PR review and CI:

- Humans can see which repo rules apply to a PR.
- Agents can quickly identify which cards to inspect before or after editing.
- CI can surface relevant cards without expanding `AGENTS.md`.
- Future verification guidance can be attached to the matched cards.

## Initial Product Shape

Prefer a narrow command that accepts changed paths explicitly:

```sh
garden check internal/cmd/root.go internal/app/app.go
```

Potential output:

```txt
Garden context for changed files:

internal/cmd/root.go
  .garden/context/app-layer-architecture.md
    scope: internal/cmd/**

internal/app/app.go
  .garden/context/app-layer-architecture.md
    scope: internal/app/**
```

The implemented command is `garden check`. Earlier candidates were:

- `garden context check`
- `garden context match`
- `garden cards match`

The important behavior is the deterministic path-to-card mapping.

## GitHub CI Use Case

A GitHub Action or CI script could collect PR changed files and call Garden:

```sh
garden check --changed-file-list changed-files.txt
```

The first useful CI mode should probably be non-blocking output:

```txt
Garden context for this PR:

internal/cmd/root.go
  read .garden/context/app-layer-architecture.md

internal/contextcard/card.go
  read .garden/context/context-card-format.md
```

Later CI modes could become stricter:

- Fail if `AGENTS.md` is stale.
- Fail if important changed files are uncovered by any card.
- Print or run verification guidance from matched cards.
- Require an explicit acknowledgement file or PR comment convention.

Avoid making uncovered-file checks blocking until the repo has mature context coverage.

## Keep AGENTS.md Small

This feature should not add long PR-specific context, verification commands, or matched-file summaries to `AGENTS.md`.

`AGENTS.md` should remain the always-visible static router. This command should generate a task-specific or PR-specific summary only when asked.

## Suggested Implementation Boundaries

Likely package responsibilities:

- `internal/contextcard`: continue owning card parsing and `scope` metadata.
- `internal/app`: orchestrate changed-path input and card matching.
- `internal/cmd`: own CLI flags, argument validation, and command UX.
- `internal/output`: format the match report.

Avoid putting matching behavior directly in `internal/cmd`.

The matching logic could live in `internal/app` if it is small, or in a focused lower-level package if glob semantics become nontrivial.

## Matching Rules To Decide

Open decisions before implementation:

- Which glob semantics should Garden use for scopes?
- Should a changed file match all matching cards, or only the most specific card?
- Should matching preserve input file order, sorted output order, or both?
- Should duplicate card matches be grouped by changed file, by card, or both?
- Should missing files still be matchable as paths, useful for deleted files in a PR?
- Should path matching normalize `./`, OS separators, and leading slashes?

Initial recommendation:

- Treat changed paths as repo-relative strings.
- Normalize paths to slash-separated relative paths.
- Match all cards whose `scope` includes the path.
- Output grouped by changed file, then card path.
- Keep deleted files matchable because CI changed-file lists include deleted paths.

## Future Verification Extension

This feature pairs naturally with card-level verification guidance.

A future card might include:

````md
## Verification

Run:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
rg '"github.com/aric/garden/internal/agents"' internal/cmd
```

Expected:

- Tests pass.
- `internal/cmd` has no direct imports of `internal/agents`.
````

Then the changed-files context check could include suggested verification for matched cards.

Do not require structured verification metadata in the first version. A conventional Markdown section is probably enough to dogfood the idea.

## Acceptance Criteria For First Version

- A command can accept changed file paths explicitly.
- Garden loads all valid context cards.
- Garden maps changed paths to matching cards using card `scope` values.
- Output is deterministic and stable.
- The command does not modify files.
- Existing `AGENTS.md` sync behavior is unchanged.
- Full test suite passes:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
```

## Test Ideas

Add focused tests for:

- One changed file matching one card.
- One changed file matching multiple cards.
- Multiple changed files matching the same card.
- Changed file matching no cards.
- Stable output regardless of card load order.
- Slash normalization for paths.
- Deleted-file style paths that do not exist on disk.

If matching semantics become subtle, add property-based tests for order independence and deterministic output.

## Non-Goals For First Version

- Do not execute verification commands.
- Do not parse complex Markdown verification sections.
- Do not post GitHub PR comments directly from Garden.
- Do not add a full GitHub Action wrapper yet.
- Do not make uncovered files fail CI by default.
- Do not expand the generated `AGENTS.md` block.

## Suggested First Slice

Build the local command first:

```sh
garden check internal/cmd/root.go
```

Once that is solid, add file-list input for CI:

```sh
garden check --changed-file-list changed-files.txt
```

Only after dogfooding the output should Garden consider stricter CI modes.
