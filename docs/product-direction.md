# Garden Product Direction

Garden is a PR context and verification reporter for agent-assisted software work.

Its job is to answer, at review time:

```txt
What repo constraints apply to this change?
What evidence should prove they were preserved?
Did this PR also change the evidence?
```

## Core Thesis

Rule files are advisory constraints. Tests and analysis are physical constraints. Garden should connect the two at the moment where they matter most: PR review.

Garden should not try to make agents obey prose rules. It should make the relevant rules, checks, and changed verification surfaces visible to humans and CI.

The narrow product promise:

```txt
Garden turns changed files into reviewable constraint evidence.
```

## Why This Exists

Agent guidance files help with discovery, but they do not guarantee behavior. Agents can see a rule and still violate it when implementation constraints pile up.

The useful tool is not another AI reviewer. The useful tool is deterministic review context:

```txt
changed files
-> scoped repo guidance
-> suggested verification
-> verification-surface warnings
```

That gives reviewers a compact way to ask:

- Which local repo rules apply here?
- Which checks are supposed to enforce those rules?
- Did the PR weaken tests, CI, lint config, build scripts, or verification guidance?

## Product Shape

Garden should grow into a GitHub PR reporter backed by a small local CLI.

In a PR, Garden should produce a compact summary:

```txt
Garden review context

Changed:
  internal/cmd/root.go

Relevant constraint:
  .garden/context/app-layer-architecture.md
  matched: internal/cmd/**

Suggested verification:
  env GOCACHE=/tmp/garden-go-build go test ./...
  rg '"github.com/aric/garden/internal/agents"' internal/cmd

Verification surfaces changed:
  none
```

The first implemented slice is the local preview command. It exists primarily as the engine behind the later PR reporter:

```sh
garden check --changed internal/cmd/root.go
```

Later CI-oriented input modes should build on the same report engine:

```sh
garden check --changed-file-list changed-files.txt
garden check --git-diff
```

## Data Model

Garden context should stay in normal Markdown files.

`AGENTS.md` remains the always-visible router for agents before they edit. Context cards become the shared data layer for both agent routing and PR review.

```txt
AGENTS.md = compact before-coding router
.garden/context/*.md = scoped guidance and verification notes
garden check = changed-files to review evidence
```

Do not put long explanations, PR-specific summaries, or verification command lists into `AGENTS.md`.

For agent workflows, `AGENTS.md` is the discovery layer. For review and CI workflows, `garden check` generates task-specific summaries on demand.

Cards should stay readable without Garden:

````md
---
scope:
  - internal/cmd/**
---

# App Layer Architecture

`internal/cmd` owns Cobra command wiring.
Commands should not bypass `internal/app` to call lower-level packages directly.

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

The `## Verification` section should start as a Markdown convention. Do not add a schema until dogfooding proves the need.

## Verification Surface Awareness

Checks only act like physical constraints if they cannot be quietly weakened.

Garden should call out changed verification surfaces, especially:

- Test files.
- CI workflows.
- Lint and formatting configs.
- Build scripts.
- Garden context cards.
- Verification sections inside cards.

A PR that changes implementation and its checks is not automatically wrong. It should be higher-attention.

CI should prefer approved checks over arbitrary commands copied from an untrusted PR.

Garden is not the security boundary. Branch protection, code owners, CI permissions, and human review own enforcement. Garden's job is to make weakening the boundary obvious.

## First Version

The first version should be boring and deterministic.

Build:

- Changed-file input.
- Deterministic scope matching from context cards.
- Compact text output.
- Suggested verification extraction from `## Verification` sections.
- Verification-surface warnings.
- Local report-only CLI usage.

Build next:

- Report-only GitHub CI usage.
- `--changed-file-list` input.
- `--git-diff` input.

Scope matching should use card `scope` globs, not LLM judgment. The same changed files should always produce the same card list.

Do not build yet:

- AI judgment about relevance.
- Automatic execution of arbitrary verification commands.
- Blocking uncovered files by default.
- A full policy engine.
- A custom editor or card-editing workflow.

## Success Criteria

Garden is worth continuing if its PR summary makes review faster and less dependent on memory.

For this repo, the first dogfood test is:

```txt
When a PR changes internal/cmd/root.go,
Garden points to app-layer-architecture.md,
shows the relevant checks,
and flags if tests or context cards changed too.
```

If the summary is noisy, vague, or ignored, stop and simplify.

## Existing Commands

Keep the current foundation:

```txt
garden init
garden new
garden remove
garden agents sync
garden lint
```

The next command should serve the PR reporter:

```txt
garden check
```

Avoid adding commands that wrap normal Markdown browsing or editing. Context cards should remain files that humans and agents can read and edit with ordinary tools.

Avoid expanding Garden into a general docs manager, memory system, or AI review product.

## Sources And Influences

- Constraint decay paper: https://arxiv.org/abs/2605.06445
- Bob Martin response: rule files are advisory; tests and analysis are physical constraints.
- CODEOWNERS: deterministic changed-file-to-reviewer routing.
- Danger: deterministic PR-time automation.
- AGENTS.md and agent rules: useful for before-coding discovery, insufficient as enforcement.
