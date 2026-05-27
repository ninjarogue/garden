---
scope:
  - README.md
  - docs/product-direction.md
tags:
  - product
---

# Product Direction

Garden exists to keep agent-facing repo context small, discoverable, reviewable, and tied to verification evidence.

Core model:

```txt
AGENTS.md = small always-visible router
.garden/context/*.md = human-readable context cards
garden check = changed-files to review evidence
```

Preserve the static `AGENTS.md` router and Markdown context-card workflow. Product decisions should also help reviewers connect changed files to relevant constraints, suggested verification, and changed verification surfaces.

Editing context cards directly is part of the product model. Garden should validate and sync Markdown cards, not wrap normal file editing in another command.

Product surface should stay narrow:

- `garden init`
- `garden new`
- `garden remove`
- `garden agents sync`
- `garden lint`
- `garden check`

When changing product docs or README examples, preserve the idea that agents discover context from `AGENTS.md`, then read matching `.garden/context/*.md` files with normal file tools. `garden check` should generate task-specific review summaries on demand rather than expanding `AGENTS.md`.
