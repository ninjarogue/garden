---
scope:
  - README.md
  - docs/product-direction.md
tags:
  - product
---

# Product Direction

Garden exists to keep agent-facing repo context small, discoverable, and reviewable.

Core model:

```txt
AGENTS.md = small always-visible router
.garden/context/*.md = human-readable context cards
```

Prefer improving the static `AGENTS.md` router and Markdown context-card workflow. Product decisions should reinforce that core loop.

Editing context cards directly is part of the product model. Garden should validate and sync Markdown cards, not wrap normal file editing in another command.

Initial product surface should stay narrow:

- `garden init`
- `garden new`
- `garden remove`
- `garden agents sync`
- `garden lint`

When changing product docs or README examples, preserve the idea that agents discover context from `AGENTS.md`, then read matching `.garden/context/*.md` files with normal file tools.
