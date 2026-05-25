---
scope:
  - AGENTS.md
  - internal/agents/**
  - docs/compact-index-syntax.md
tags:
  - agents
---

# Agents Router

`AGENTS.md` is generated from context-card frontmatter. Do not hand-edit the Garden-managed section between `<!-- garden:agents:start -->` and `<!-- garden:agents:end -->`.

To change the index:

1. Edit the relevant `.garden/context/*.md` card frontmatter.
2. Run `go run ./cmd/garden agents sync --apply`.
3. Run `go run ./cmd/garden lint`.

`internal/agents` owns marker parsing, compact index rendering, sync behavior, and AGENTS lint findings. Changes there should preserve deterministic output and avoid rewriting human-authored content outside Garden markers.

`docs/compact-index-syntax.md` documents the compact index format. Keep generated output aligned with that document and avoid adding Garden-specific compact vocabulary unless it is explicitly approved there.
