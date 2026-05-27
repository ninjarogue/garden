---
scope:
  - internal/app/**
  - internal/cmd/**
  - internal/output/**
tags:
  - architecture
---

# App Layer Architecture

The app layer coordinates storage, AGENTS rendering, and CLI output without owning domain-specific parsing or rendering details.

Key boundaries:

- `internal/cmd` owns Cobra command wiring, flags, argument validation, stdout/stderr wiring, and command-level errors.
- `internal/app` owns use-case orchestration and app-owned DTOs/interfaces.
- `internal/contextcard` owns Markdown card storage, parsing, validation, and template rendering.
- `internal/agents` owns AGENTS.md marker logic, compact index rendering, sync behavior, and lint findings.
- `internal/output` owns human-readable CLI output formatting.

When changing command behavior, keep command tests focused on CLI UX and delegate business behavior to `internal/app` or lower packages.

When changing `internal/app`, preserve dependency injection through `CardStore` and `AgentsFile` so behavior can be tested without shelling out.

Avoid adding cross-package shortcuts from commands directly into `internal/agents` or `internal/contextcard` unless the app layer has no orchestration value for that path.

## Verification

Run:

```sh
env GOCACHE=/tmp/garden-go-build go test ./...
rg '"github.com/aric/garden/internal/(agents|contextcard|review)"' internal/cmd internal/output --glob '!**/*_test.go'
```

Expected:

- Tests pass.
- The `rg` command returns no production-file matches.
- `internal/cmd` only wires CLI behavior through `internal/app` and `internal/output`.
- `internal/output` formats app-owned DTOs and does not import lower-level domain packages.
