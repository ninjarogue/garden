# Garden

Garden maintains a small `AGENTS.md` router for coding agents, stores detailed repo context in Markdown cards, and reports review context for changed files.

Core model:

```txt
AGENTS.md = small always-visible router
.garden/context/*.md = human-readable context cards
garden check = changed-files to review evidence
```

Agents discover relevant cards from `AGENTS.md`, then read the Markdown card files with normal file tools.
Reviewers can run `garden check` to see which cards, suggested verification, and verification-surface warnings apply to changed files.

## Commands

Initialize context-card storage:

```sh
garden init
```

Create a context card:

```sh
garden new routes-query-modules --scope "src/routes/**" --tag database
```

Sync the generated Garden section in `AGENTS.md`:

```sh
garden agents sync --apply
```

Preview the same sync without writing:

```sh
garden agents sync
```

Validate context cards and the `AGENTS.md` index:

```sh
garden lint
```

Report review context for changed files:

```sh
garden check --changed internal/cmd/root.go
```

Remove a context card:

```sh
garden remove routes-query-modules
```

Run `garden agents sync --apply` again after adding, editing, or removing cards.

Edit context cards directly as Markdown files; Garden validates and syncs them.

## Context Cards

Cards live in `.garden/context/*.md` and use small YAML frontmatter:

```md
---
scope:
  - src/routes/**
tags:
  - database
  - tenant-scoping
---

# Routes Query Modules

Route files should use query modules for database access.
```

Required fields:

- `scope`: one or more repo-relative globs

Optional fields:

- `tags`: labels for human organization

## Verification

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
```
