# Constraint Decay And Garden

This note captures a working take after reading arXiv:2605.06445 and comparing it to Garden's current product direction.

Paper:

- https://arxiv.org/abs/2605.06445

## Core Take

Garden is good foundation work, but the paper's lesson is that advisory context is not enough. The next serious version should connect scoped context to verification.

Garden should be pitched as context hygiene and routing for agents, not agent memory or agent correctness.

If Garden grows toward lightweight structural verification, it becomes much more defensible in the world this paper describes.

## Why This Matters

The paper argues that agents degrade as structural constraints accumulate. They may see the rule, understand the rule locally, and still violate it when the implementation requires coordinating architecture, persistence, tests, and conventions.

That maps directly onto repo guidance:

- A small `AGENTS.md` router helps agents discover relevant context.
- Scoped context cards help keep detailed guidance local and reviewable.
- But reading a card does not prove the resulting edit followed the card.

So Garden should not stop at "changed files imply relevant cards." It should help answer: "What checks would give us confidence that this card's guidance survived the edit?"

## Priority Ideas

1. Add a deterministic changed-files-to-relevant-cards check for humans and CI, probably without bloating `AGENTS.md`.
2. Let cards carry explicit verification guidance, even just a conventional `## Verification` section at first.
3. Add lint for obvious context-system failure modes: duplicate scopes, overlapping or conflicting cards, orphaned cards, and uncovered important directories.
4. Keep resisting product sprawl. The narrow command surface is a strength.

## Example Direction

A card should not only say:

> When editing `internal/app/**`, preserve the app layer boundary.

It should also say how to check that boundary:

````md
## Verification

Run:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
rg '"github.com/aric/garden/internal/agents"' internal/cmd
rg '"github.com/aric/garden/internal/contextcard"' internal/cmd
```

Expected:

- Tests pass.
- `internal/cmd` has no direct imports of `internal/agents` or `internal/contextcard`.
````

Garden could then report:

```txt
Changed files:
  internal/cmd/root.go

Relevant card:
  .garden/context/app-layer-architecture.md

Suggested verification:
  env GOCACHE=/tmp/garden-go-build go test ./...
  rg '"github.com/aric/garden/internal/agents"' internal/cmd
```

## Product Framing

Garden's strongest idea is not making agents smarter. It is making repo constraints easier to discover, review, and verify.

The product promise should stay narrow:

```txt
Garden keeps agent-facing repo context small, scoped, and checkable.
```

## Open Questions

- Should verification guidance be just Markdown convention first, or structured frontmatter?
- Should Garden execute checks, or only surface the relevant checks?
- Should changed-file matching use git diff by default, explicit paths, or both?
- How should Garden handle overlapping scopes without becoming noisy?
- What is the smallest CI command that proves the model?
