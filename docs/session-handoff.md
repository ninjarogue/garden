# Session Handoff

## 2026-05-25 09:19 CST

### Current State

Branch: `master`

Working tree was clean before this handoff update. The branch is ahead of `origin/master`.

Recent commits:

- `11ba866 docs: update session log`
- `1c42150 ref: decouple app layer types`

Latest verification for the refactor:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
```

Result: all packages passed.

### Product Direction

Garden should prove the simple AGENTS router workflow before adding new commands or lint rules.

Core model:

```txt
AGENTS.md = small always-visible router
.garden/context/*.md = human-readable context cards
garden = authoring, indexing, syncing, and linting tool
```

Garden should not center runtime retrieval, context packing, ranking, or session injection right now.

### Recommended Next Move

Dogfood Garden in this repo.

Steps:

1. Run `garden init`.
2. Add a few real context cards:
   - `product-direction`
   - `context-card-format`
   - `app-layer-architecture`
   - `testing-guidelines`
3. Run `garden agents sync --apply`.
4. Run `garden lint`.

This should prove whether the core workflow is actually pleasant and whether the current card format and AGENTS index are enough.

### What To Avoid For Now

Do not add new product surface before dogfooding.

Avoid adding lint rules just because they are mentioned as possible future checks. Future checks like duplicate, conflicting, missing, broad, or orphaned context should only be added when dogfooding exposes a concrete pain point.

Keep lint objective and low-noise.

### Likely Follow-Up After Dogfooding

If dogfooding exposes real problems, improve the smallest relevant part:

- Card format if writing cards feels awkward.
- AGENTS index wording if discovery is unclear.
- Lint if the workflow allows stale or misleading context.
- Documentation if users cannot understand the loop quickly.

Prefer improving the existing core loop over adding commands like `list`, `search`, `edit`, `pack`, or `read`.

### Important Preferences

- Use TDD for behavior changes.
- Keep changes surgical.
- Do not commit without explicit approval.
- Keep responses concise.
