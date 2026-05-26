# Garden

Garden maintains a small `AGENTS.md` router for coding agents and stores detailed repo context in Markdown cards.

Core model:

```txt
AGENTS.md = small always-visible router
.garden/context/*.md = human-readable context cards
garden = authoring, indexing, syncing, and linting tool
```

Agents discover relevant cards from `AGENTS.md`, then read the Markdown card files with normal file tools.

## Running Garden

During development, run Garden from the repository root with `go run`:

```sh
go run ./cmd/garden <command>
```

For example:

```sh
go run ./cmd/garden init
go run ./cmd/garden lint
```

To build a local binary:

```sh
go build -o garden ./cmd/garden
./garden <command>
```

To install Garden as a terminal command:

```sh
go install github.com/aric/garden/cmd/garden@latest
```

Make sure Go's binary directory is on your `PATH`. It is usually:

```sh
export PATH="$HOME/go/bin:$PATH"
```

After installing, run Garden from the root of the project you want to manage:

```sh
garden <command>
```

Running `garden` without a command prints the command help and exits.

## First-Time Workflow

Use this flow when adding Garden to a repository for the first time:

```sh
garden init
garden new project-overview --scope "**/*" --tag overview
garden agents sync
garden agents sync --apply
garden lint
```

The first `agents sync` previews the generated `AGENTS.md` change. Re-run with `--apply` when the preview looks correct.

After creating the card, edit the generated Markdown file in `.garden/context/` and replace the placeholder body with real project context.

## Existing-Project Workflow

Use this flow after Garden is already set up:

```sh
garden new routes-query-modules --scope "src/routes/**" --tag database
garden agents sync
garden agents sync --apply
garden lint
```

When you edit, add, or remove cards, run `garden agents sync --apply` again so `AGENTS.md` stays current.

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
