---
scope:
  - '**/*_test.go'
  - docs/testing.md
tags:
  - testing
---

# Testing Guidelines

Garden tests should make behavior failures easy to localize.

Run the full suite with:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
```

Test shape:

- Keep command tests focused on one CLI behavior: `init`, `new`, `agents sync`, `lint`, or `remove`.
- Keep output tests focused on one formatting contract: preview diff, findings, applied/no-change, lint pass, or lint findings.
- Use temp directories and real file reads/writes for app, command, and context-card behavior.
- Use exact equality for stable generated strings such as AGENTS blocks, compact indexes, card templates, and CLI output.
- Use substring assertions only for intentionally flexible error messages or defensive checks.
- Add parser/store tests before changing context-card validation.
- Add AGENTS render/lint tests before changing compact index syntax or marker behavior.

Property-based tests live in `*_property_test.go` files near the package they protect. Keep example tests for readable edge cases and use PBT for invariants such as round trips, idempotence, ordering, and preservation.
