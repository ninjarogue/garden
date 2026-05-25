# Garden Product Direction

Garden improves the `AGENTS.md` workflow for coding agents.

The product should stay close to the pattern that already works: agents read `AGENTS.md`, discover compact repo guidance, and use normal file-reading tools when they need detail. Garden should not make the default workflow depend on injecting large retrieved context into the session.

## Core Bet

Agent context should live in the repo as readable, reviewable files.

`AGENTS.md` should stay small. It should route agents to the right context, not contain every rule, exception, warning, and workflow note.

Garden's job is to maintain that system:

```txt
AGENTS.md = small always-visible router
context cards = human-readable source of truth
garden = authoring, indexing, syncing, and linting tool
```

## Product Identity

Garden is not primarily a runtime retrieval engine.

Garden is a maintainer for agent-readable repo context.

It should help teams:

- Keep `AGENTS.md` compact and useful.
- Store detailed context in Markdown cards.
- Keep those cards discoverable from `AGENTS.md`.
- Create and edit cards with consistent metadata.
- Detect stale, duplicate, conflicting, missing, overly broad, or orphaned context.

## Source Of Truth

The source of truth should move toward Markdown context cards:

```txt
.garden/context/
  routes-query-modules.md
  retrieval-ranking.md
  product-direction.md
```

Each card should be readable without Garden:

```md
---
scope:
  - src/routes/**
tags:
  - database
  - tenant-scoping
---

Route files should not import DB clients directly.

Use query modules instead. They centralize tenant scoping, audit logging,
and permission checks.
```

The Markdown body is the repo knowledge. The frontmatter is indexing metadata.

## AGENTS.md Router

Garden should manage a compact generated section inside `AGENTS.md`:

```md
### Garden Context

Detailed agent context lives in `.garden/context/*.md`.

Before editing a listed area, inspect the matching context card.

Index:
[Garden Context Index]|root:.garden/context
|IMPORTANT:Before editing a listed area, inspect the matching context card
|src/routes/**:.garden/context/routes-query-modules.md
|internal/contextcard/**:.garden/context/context-card-format.md
|**/*:.garden/context/product-direction.md
```

The agent sees that Garden exists, which files have guidance, and which card to open. The agent then uses ordinary file-reading tools to inspect the card only when needed.

Garden should not ask agents to run a second discovery command when `AGENTS.md` already contains the map. For agent workflows, `AGENTS.md` is the discovery layer.

## Deferred Runtime Retrieval

`garden pack` and `garden read` should not be central to the product direction right now.

They risk recreating the original problem by injecting repeated or oversized context into the session. They may be useful later as debug or experimental commands, but Garden should first prove value by improving the static `AGENTS.md` and context-card workflow.

## Initial Command Shape

The likely core commands are:

```txt
garden init
garden new
garden remove
garden agents sync
garden lint
```

`garden new <slug>` should create a Markdown context card template in `.garden/context/<slug>.md`.

Example:

```txt
garden new routes-query-modules --scope "src/routes/**" --tag database
```

Generated card:

```md
---
scope:
  - src/routes/**
tags:
  - database
---

# Routes Query Modules

Write the repo context here.
```

The file path/slug is the card identity. No separate `id` field is needed.

`garden remove` should delete a context card. `garden agents sync` should then remove it from the `AGENTS.md` router/index.

`garden agents sync` should update the `AGENTS.md` router/index.

`garden lint` should protect the quality of the context system.

Editing should happen directly in the Markdown card. Humans can use their editor. Agents can use normal file-edit tools. Garden should validate and sync the result, not wrap every edit in a CLI.

The required card metadata should stay small:

```yaml
scope:
  - src/routes/**
```

Optional tags can provide labels for human organization:

```yaml
tags:
  - database
  - tenant-scoping
```

`priority` is not part of the new core model. It belonged to runtime retrieval ranking, and Garden is no longer centering task-specific ranking.

`garden edit`, `garden list`, and `garden search` are not part of the initial product shape. If Markdown cards are the source of truth, normal shell/editor tools such as `ls`, `rg`, file reads, and file edits are already good enough. Garden's unique value is structure, `AGENTS.md` sync, and context health.

## Lint Scope

`garden lint` should start as a basic correctness check. It should only enforce what Garden needs in order to trust the context-card system and regenerate the `AGENTS.md` index.

Initial checks:

- Each `.garden/context/*.md` file has YAML frontmatter.
- `scope` exists and has at least one non-empty glob.
- `scope` does not contain `CHANGE_ME`.
- `tags`, if present, is a list.
- Index metadata does not contain compact-index delimiters or control characters that would make `AGENTS.md` unsyncable.
- The Markdown body is non-empty.
- The card filename/slug is valid.
- `AGENTS.md` has the Garden managed block.
- The `AGENTS.md` index matches the current context cards.

Avoid subjective lint rules in the first version. Do not lint vague tags, secret-like words, or "too many scopes" until there is a concrete product reason and a low-false-positive rule.

## Vercel-Inspired Direction

Vercel's public guidance supports the router approach:

- Repo guidance files like `AGENTS.md`, `CLAUDE.md`, `.cursorrules`, and `.github/copilot-instructions.md` can guide agent behavior.
- A compact `AGENTS.md` index can outperform approaches that hide guidance elsewhere, because the agent reliably sees it.
- Full reference material can live outside `AGENTS.md`, with the always-visible file pointing agents to the right place.

Garden should push that pattern further: keep `AGENTS.md` small, make detailed context card-based, and maintain the index automatically.

## Sources

- Vercel changelog: https://vercel.com/changelog/vercel-agent-code-reviews-now-follow-your-code-guidelines
- Vercel blog on `AGENTS.md` evals: https://vercel.com/blog/agents-md-outperforms-skills-in-our-agent-evals
- Vercel agent readability spec: https://vercel.com/kb/guide/agent-readability-spec
