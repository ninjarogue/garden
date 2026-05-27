# Archived Garden Product Direction: Context Verification Thesis

Archived on 2026-05-26 before replacing the product direction with a PR-context reporter thesis.

---

# Garden Product Direction

Garden keeps agent-facing repo context small, scoped, and checkable.

The core thesis:

```txt
Rule files are advisory constraints.
Tests and analysis are physical constraints.
Garden should connect the two.
```

Garden should not try to make agents obey rules. It should route agents to the relevant rules, then route humans and CI to the checks that make those rules real.

Those checks only matter if they cannot be quietly weakened. Garden should also make changes to tests, CI, lint configs, build scripts, and other verification surfaces visible for review.

## Product Bet

Agent guidance is useful, but it is not enforcement.

Agents can forget, misread, or rationalize around written rules. The durable constraints are tests, static checks, property tests, mutation tests, architectural checks, and other verification that runs outside the agent's immediate narrative.

Garden's job is to connect advisory context to those physical constraints:

```txt
AGENTS.md = small always-visible router
context cards = scoped repo guidance
verification = concrete checks tied to that guidance
garden = sync, lint, matching, and verification routing
```

## Product Identity

Garden is a context hygiene and verification-routing tool for agent-assisted software work.

It should help teams:

- Keep `AGENTS.md` compact and useful.
- Store detailed repo guidance in scoped Markdown cards.
- Match changed files to relevant cards deterministically.
- Surface verification guidance tied to those cards.
- Flag changes to the verification surfaces that make constraints enforceable.
- Keep generated agent context synced and reviewable.
- Detect stale, missing, duplicate, conflicting, overly broad, or orphaned context.

## The Workflow

The default loop should stay simple:

```txt
1. AGENTS.md routes the agent to relevant context cards.
2. Context cards explain the local rules and expectations.
3. Garden maps changed files back to relevant cards.
4. Garden highlights any changed verification surfaces.
5. Humans, agents, or CI run the checks tied to those cards.
```

The important shift is that a card should not only explain a rule; it should help prove whether the rule survived the edit. Garden should also make it obvious when a PR changes the checks that provide that proof.

## Context Cards

Context cards remain normal Markdown files in `.garden/context/*.md`.

Required frontmatter should stay small:

```yaml
scope:
  - internal/app/**
```

The body should stay human-readable and operational:

````md
# App Layer Architecture

`internal/cmd` owns Cobra command wiring.
`internal/app` owns use-case orchestration.
Avoid commands calling `internal/agents` directly.

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

The `## Verification` section can start as convention, not structured metadata. Garden should first prove that surfacing this guidance is useful before adding a schema.

## AGENTS.md Router

`AGENTS.md` should remain a compact always-visible router:

```txt
AGENTS.md = small map
.garden/context/*.md = detailed guidance
```

Do not put long explanations, PR-specific summaries, or verification command lists into `AGENTS.md`.

For agent workflows, `AGENTS.md` is the discovery layer. For review and CI workflows, Garden commands can generate task-specific summaries on demand.

## Changed-Files Context Check

The next major product direction is deterministic changed-file matching:

```txt
Given these changed files, which Garden context cards apply?
```

Example:

```sh
garden context check --changed internal/cmd/root.go
```

Potential output:

```txt
Garden context for changed files:

internal/cmd/root.go
  .garden/context/app-layer-architecture.md
    scope: internal/cmd/**
```

This should use card `scope` globs, not LLM judgment. The same changed files should always produce the same card list.

## CI Direction

Garden should be useful in GitHub CI without requiring a large platform integration.

A first CI use could collect PR changed files and print the relevant cards:

```txt
Garden context for this PR:

internal/cmd/root.go
  read .garden/context/app-layer-architecture.md
```

Early CI behavior should be conservative:

- Fail when generated Garden state is stale.
- Print relevant cards for review.
- Avoid failing on uncovered files until a repo has mature context coverage.
- Avoid executing arbitrary verification commands until the trust and safety model is clearer.

Later CI behavior can become stricter:

- Warn or fail on important uncovered paths.
- Surface verification sections from matched cards.
- Run approved checks tied to matched cards.

## Verification Trust Model

Tests and checks are only physical constraints if the agent cannot quietly weaken them.

Garden should treat verification files as a higher-trust surface:

- Test files, CI workflows, lint configs, build scripts, and verification sections should be visible in Garden's changed-file reports.
- Changes to verification surfaces should be called out explicitly for human review.
- CI should prefer approved checks over arbitrary commands copied from an untrusted PR.
- Branch protection and code ownership should protect CI workflows and critical test infrastructure.
- A PR that changes both implementation and the checks that enforce it is not automatically wrong, but it should be marked as higher attention.

Garden does not need to become a security boundary. It should make weakening the boundary obvious.

## Command Surface

The core command surface should remain narrow:

```txt
garden init
garden new
garden remove
garden agents sync
garden lint
```

New commands should be added only when they serve the context-to-verification loop. The likely next candidate is a changed-files context command, but the exact name is still open.

Avoid adding commands that wrap normal Markdown browsing or editing. Context cards should remain files that humans and agents can read and edit with ordinary tools.

## Non-Goals

Garden should not promise that agents will obey rules.

Garden should not become:

- An agent memory system.
- A replacement for tests or CI.
- A general policy engine.
- A wrapper around normal file editing.
- A place to hide large guidance outside reviewable Markdown.

Garden should make repo constraints easier to discover, review, and verify.

## Sources And Influences

- Constraint decay paper: https://arxiv.org/abs/2605.06445
- Bob Martin response: rule files are advisory; tests and analysis are physical constraints.
- Vercel agent readability work: compact always-visible repo guidance can improve agent behavior.
