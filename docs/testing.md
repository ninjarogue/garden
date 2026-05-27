# Garden Test Suite

Garden tests should make behavior failures easy to localize.

## Verification

Run the full suite with:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
```

## Test Shape

- Keep command tests focused on one CLI behavior: `init`, `new`, `agents sync`, `lint`, `remove`, or `check`.
- Keep output tests focused on one formatting contract: preview diff, findings, applied/no-change, lint pass, lint findings, or check reports.
- Use temp directories and real file reads/writes for app, command, and context-card behavior.
- Use exact equality for stable generated strings in the package that owns them, such as AGENTS blocks, compact indexes, card templates, and output formatting.
- In command tests, assert command UX and side effects without duplicating full generated payloads already covered by lower package tests.
- Use substring assertions only for intentionally flexible error messages or defensive checks.
- Add parser/store tests before changing context-card validation.
- Add AGENTS render/lint tests before changing compact index syntax or marker behavior.

## Current Focus Areas

- `internal/contextcard`: Markdown frontmatter parsing, card template rendering, duplicate prevention, YAML-sensitive glob scopes, and unsyncable metadata rejection.
- `internal/agents`: compact index rendering, marker upserts, sync behavior, and lint findings.
- `internal/cmd`: Cobra command wiring and CLI UX.
- `internal/app`: orchestration across card storage and AGENTS sync/lint.
- `internal/output`: stable human-readable command output.
- `internal/review`: deterministic changed-file matching, verification extraction, and verification-surface warnings.
