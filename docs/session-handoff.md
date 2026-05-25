# Session Handoff

## 2026-05-25 09:57 CST

### Current State

Branch: `master`

Working tree was clean before this handoff update. The branch is ahead of `origin/master`.

Recent commits before this update:

- `138ee80 docs: update handoff notes`
- `3a8ad67 docs: add property-based testing findings`
- `fba97f5 test: add property-based coverage`
- `11ba866 docs: update session log`

Latest verification after adding PBT tests:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
```

Result: all packages passed.

PBT work is committed:

- Real property tests were added for `RenderIndex` ordering and context-card template/parse round trips.
- Findings for a future PBT skill live in `docs/property-based-testing-findings.md`.
- Next PBT targets live in `docs/pbt-next-targets-handoff.md`.

### Product Direction

Garden should prove the simple AGENTS router workflow before adding new commands or lint rules.

Core model:

```txt
AGENTS.md = small always-visible router
.garden/context/*.md = human-readable context cards
garden = authoring, indexing, syncing, and linting tool
```

Garden should keep improving the `AGENTS.md` router and context-card loop.

### Dogfooding State

Dogfooding is in progress in this repo.

Completed:

- `garden init` created `.garden/context`.
- Context cards were created with `garden new`.
- Card bodies/frontmatter were edited directly as Markdown.
- `garden agents sync --apply` regenerated `AGENTS.md`.
- `garden lint` passed.

Current finding: direct Markdown edits plus sync/lint are enough for the card-edit workflow. Do not add a card-editing command unless later dogfooding exposes a concrete gap.

### What To Avoid For Now

Do not add new product surface before the current dogfooding changes are reviewed.

Avoid adding lint rules just because they are mentioned as possible future checks. Future checks like duplicate, conflicting, missing, broad, or orphaned context should only be added when dogfooding exposes a concrete pain point.

Keep lint objective and low-noise.

### Likely Follow-Up After Dogfooding

If dogfooding exposes real problems, improve the smallest relevant part:

- Card format if writing cards feels awkward.
- AGENTS index wording if discovery is unclear.
- Lint if the workflow allows stale or misleading context.
- Documentation if users cannot understand the loop quickly.

Prefer improving the existing core loop over adding more commands.

### Important Preferences

- Use TDD for behavior changes.
- Keep changes surgical.
- Do not commit without explicit approval.
- Keep responses concise.
